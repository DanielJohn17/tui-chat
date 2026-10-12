#!/usr/bin/env node

/**
 * sync-release-sanity.js
 * Synchronizes GitHub Releases metadata, changelog markdown, and binary
 * download matrices to Sanity.io Content Lake for the LineTalk Landing Page.
 *
 * Implements non-destructive upserts (createIfNotExists + patch) preserving
 * manual editorial enrichments in Sanity Studio.
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

function getGitShortHash(tag) {
  try {
    if (tag) {
      return execSync(`git rev-list -n 1 ${tag} 2>/dev/null`, { encoding: 'utf8' }).trim().slice(0, 7);
    }
    return execSync('git rev-parse --short HEAD 2>/dev/null', { encoding: 'utf8' }).trim();
  } catch {
    return undefined;
  }
}

function getLatestTag() {
  try {
    return execSync('git describe --tags --abbrev=0 2>/dev/null', { encoding: 'utf8' }).trim();
  } catch {
    return 'v1.1.0';
  }
}

function parseReleaseNotes(notesContent, defaultTag) {
  let title = `LineTalk ${defaultTag}`;
  let summary = 'Production binary distribution and security updates.';
  const lines = notesContent.split('\n');

  // Look for header title: e.g. "# Release v1.1.0: Some Title" or "# Some Title"
  for (const line of lines) {
    const trimmed = line.trim();
    if (trimmed.startsWith('# ')) {
      const headerText = trimmed.replace(/^#\s+/, '');
      if (headerText.includes(':')) {
        title = headerText.split(':').slice(1).join(':').trim();
      } else {
        title = headerText;
      }
      break;
    }
  }

  // Look for summary paragraph: first substantial paragraph (>= 20 chars, not heading or bullet)
  let pastHeader = false;
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    if (trimmed.startsWith('#')) {
      pastHeader = true;
      continue;
    }
    if (pastHeader && !trimmed.startsWith('*') && !trimmed.startsWith('-') && !trimmed.startsWith('>') && !trimmed.startsWith('|') && trimmed.length > 25) {
      summary = trimmed;
      break;
    }
  }

  return { title, summary };
}

async function run() {
  const repo = process.env.REPO_NAME || process.env.GITHUB_REPOSITORY || 'DanielJohn17/tui-chat';
  let tag = process.env.TAG_NAME || process.argv[2];

  if (!tag) {
    tag = getLatestTag();
    console.log(`[Sanity Sync] No tag specified, using latest detected tag: ${tag}`);
  }

  if (!tag) {
    console.error('Error: Tag name is required.');
    process.exit(1);
  }

  // Check secrets
  const projectId = process.env.SANITY_PROJECT_ID;
  const dataset = process.env.SANITY_DATASET || 'production';
  const token = process.env.SANITY_API_TOKEN;
  const allowEmptySecrets = process.env.ALLOW_EMPTY_SECRETS === 'true';
  const dryRun = process.env.DRY_RUN === 'true';
  const forceUpdate = process.env.FORCE_UPDATE === 'true';

  if (!projectId || !token) {
    if (dryRun) {
      console.log('[Sanity Sync] DRY_RUN active without credentials. Generating preview payload...');
    } else if (allowEmptySecrets) {
      console.log('::notice::SANITY_PROJECT_ID or SANITY_API_TOKEN is not configured. Skipping Sanity CMS synchronization.');
      console.log('Configure SANITY_PROJECT_ID and SANITY_API_TOKEN in GitHub repository secrets to enable automatic landing page sync.');
      process.exit(0);
    } else {
      console.error('Error: Missing required secrets SANITY_PROJECT_ID or SANITY_API_TOKEN.');
      process.exit(1);
    }
  }

  // Determine Changelog & Metadata
  let changelog = process.env.RELEASE_BODY || '';
  let title = process.env.RELEASE_NAME || `LineTalk ${tag}`;
  let summary = 'Production binary distribution and security updates.';

  const releaseNotesFile = path.resolve(process.cwd(), `.github/releases/${tag}.md`);
  if (fs.existsSync(releaseNotesFile)) {
    console.log(`[Sanity Sync] Found release notes file at ${releaseNotesFile}`);
    const notesContent = fs.readFileSync(releaseNotesFile, 'utf8');
    if (!changelog) {
      changelog = notesContent;
    }
    const parsed = parseReleaseNotes(notesContent, tag);
    if (!process.env.RELEASE_NAME) {
      title = parsed.title;
    }
    summary = parsed.summary;
  } else if (!changelog) {
    changelog = `### 🚀 LineTalk ${tag}\nProduction binary distribution and updates.`;
  }

  const publishedDate = (process.env.PUBLISHED_AT || new Date().toISOString()).split('T')[0];
  const commitHash = process.env.TARGET_COMMITISH
    ? process.env.TARGET_COMMITISH.slice(0, 7)
    : (process.env.GITHUB_SHA ? process.env.GITHUB_SHA.slice(0, 7) : getGitShortHash(tag));

  const docId = `release-${tag.replace(/[^a-zA-Z0-9_-]/g, '-')}`;

  const releaseDoc = {
    _id: docId,
    _type: 'release',
    version: tag,
    title: title,
    releaseDate: publishedDate,
    summary: summary,
    changelog: changelog,
    isPublished: true,
    commitHash: commitHash,
    downloads: {
      linux_amd64: `https://github.com/${repo}/releases/download/${tag}/linetalk-linux-amd64`,
      linux_arm64: `https://github.com/${repo}/releases/download/${tag}/linetalk-linux-arm64`,
      darwin_arm64: `https://github.com/${repo}/releases/download/${tag}/linetalk-darwin-arm64`,
      darwin_amd64: `https://github.com/${repo}/releases/download/${tag}/linetalk-darwin-amd64`,
      windows_amd64: `https://github.com/${repo}/releases/download/${tag}/linetalk-windows-amd64.exe`,
      go_install: `go install github.com/${repo}/cmd/linetalk@${tag}`
    }
  };

  const patchSet = {
    downloads: releaseDoc.downloads,
    commitHash: releaseDoc.commitHash,
    isPublished: true
  };

  if (forceUpdate) {
    patchSet.title = releaseDoc.title;
    patchSet.summary = releaseDoc.summary;
    patchSet.changelog = releaseDoc.changelog;
    patchSet.releaseDate = releaseDoc.releaseDate;
  }

  const mutations = [
    {
      createIfNotExists: releaseDoc
    },
    {
      patch: {
        id: docId,
        set: patchSet
      }
    }
  ];

  console.log(`[Sanity Sync] Preparing payload for release ${tag} (Doc ID: ${docId})`);
  console.log(`  Title: ${releaseDoc.title}`);
  console.log(`  Summary: ${releaseDoc.summary}`);
  console.log(`  Commit: ${releaseDoc.commitHash || 'N/A'}`);
  console.log(`  Release Date: ${releaseDoc.releaseDate}`);

  if (dryRun) {
    console.log('[Sanity Sync] DRY_RUN is active. Mutation payload:');
    console.log(JSON.stringify({ mutations }, null, 2));
    return;
  }

  const payload = JSON.stringify({ mutations });
  const options = {
    hostname: `${projectId}.api.sanity.io`,
    path: `/v2024-01-01/data/mutate/${dataset}`,
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`,
      'Content-Length': Buffer.byteLength(payload)
    }
  };

  return new Promise((resolve, reject) => {
    const req = https.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => { data += chunk; });
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          console.log(`✓ Successfully synced ${tag} to Sanity (${dataset}):`, data);
          resolve(data);
        } else {
          console.error(`✗ Sanity mutation failed (HTTP ${res.statusCode}):`, data);
          process.exit(1);
        }
      });
    });

    req.on('error', (err) => {
      console.error('✗ Network request error connecting to Sanity:', err.message);
      process.exit(1);
    });

    req.write(payload);
    req.end();
  });
}

run().catch((err) => {
  console.error('Unexpected error:', err);
  process.exit(1);
});
