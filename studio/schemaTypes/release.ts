import { defineField, defineType } from 'sanity'

export const releaseType = defineType({
  name: 'release',
  title: 'Release',
  type: 'document',
  fields: [
    defineField({
      name: 'version',
      title: 'Version Tag',
      type: 'string',
      validation: (rule) => rule.required(),
      description: 'e.g. v1.1.0',
    }),
    defineField({
      name: 'title',
      title: 'Release Title',
      type: 'string',
      validation: (rule) => rule.required(),
      description: 'Human-readable release title',
    }),
    defineField({
      name: 'releaseDate',
      title: 'Release Date',
      type: 'date',
      validation: (rule) => rule.required(),
    }),
    defineField({
      name: 'summary',
      title: 'Short Summary',
      type: 'text',
      rows: 3,
      description: 'Brief 1-2 sentence overview shown in release cards.',
    }),
    defineField({
      name: 'changelog',
      title: 'Changelog (Markdown)',
      type: 'text',
      rows: 15,
      description: 'GitHub-flavored markdown changelog notes.',
    }),
    defineField({
      name: 'isPublished',
      title: 'Is Published',
      type: 'boolean',
      initialValue: true,
      description: 'Only published releases are rendered on the landing page.',
    }),
    defineField({
      name: 'commitHash',
      title: 'Commit Hash',
      type: 'string',
      description: 'Git 7-character commit short hash (e.g. 1b11545)',
    }),
    defineField({
      name: 'downloads',
      title: 'Download Binaries',
      type: 'object',
      fields: [
        defineField({ name: 'linux_amd64', title: 'Linux x86_64 URL', type: 'url' }),
        defineField({ name: 'linux_arm64', title: 'Linux ARM64 URL', type: 'url' }),
        defineField({ name: 'darwin_arm64', title: 'macOS Apple Silicon (ARM64) URL', type: 'url' }),
        defineField({ name: 'darwin_amd64', title: 'macOS Intel (x86_64) URL', type: 'url' }),
        defineField({ name: 'windows_amd64', title: 'Windows x64 URL', type: 'url' }),
        defineField({ name: 'go_install', title: 'Go Install Command', type: 'string' }),
      ],
    }),
  ],
  preview: {
    select: {
      title: 'version',
      subtitle: 'title',
    },
  },
})
