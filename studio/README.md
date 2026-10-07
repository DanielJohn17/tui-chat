# LineTalk Sanity Studio (`linetalk-studio`)

Editorial Content Management Hub for LineTalk releases, release notes, and binary downloads.

## 1. Quick Setup & Local Development

```bash
cd studio
npm install
npm run dev
```

The Studio will launch locally at `http://localhost:3333`.

## 2. Environment Configuration

Create `.env.local` inside `studio/` (or configure in Vercel project settings):

```env
SANITY_STUDIO_PROJECT_ID=your_sanity_project_id
SANITY_STUDIO_DATASET=production
```

## 3. Deploying to Vercel (Hobby Free Tier)

1. Push this directory or create a new GitHub repository (`github.com/DanielJohn17/linetalk-studio`).
2. In Vercel, click **Add New Project** and select the repository.
3. Framework Preset: **Other**
   - Build Command: `npm run build`
   - Output Directory: `dist`
4. Add Environment Variables:
   - `SANITY_STUDIO_PROJECT_ID` = your Sanity Project ID
   - `SANITY_STUDIO_DATASET` = `production`
5. In Sanity Management Console ([sanity.io/manage](https://sanity.io/manage)):
   - Go to your Project -> **API** -> **CORS Origins**
   - Add your Vercel deployment URL (e.g. `https://linetalk-studio.vercel.app`) with **Allow credentials** checked.
