#!/usr/bin/env node

import { chromium } from 'playwright-extra';
import StealthPlugin from 'puppeteer-extra-plugin-stealth';
import fs from 'node:fs';

function arg(name, fallback) {
  const idx = process.argv.indexOf(name);
  if (idx === -1 || idx + 1 >= process.argv.length) return fallback;
  return process.argv[idx + 1];
}

const profile = arg('--profile', 'gmbadass');
const limit = Number(arg('--limit', '20'));
const headless = arg('--headless', 'true') !== 'false';
const storageStatePath = arg('--storage-state', '');
const sessionID = arg('--sessionid', '');

const url = `https://www.tiktok.com/@${profile}`;
chromium.use(StealthPlugin());

const launchArgs = ['--disable-blink-features=AutomationControlled'];
if (headless) {
  launchArgs.push('--headless=new');
}
const browser = await chromium.launch({
  headless,
  args: launchArgs,
});
const contextOptions = {
  userAgent:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36',
  locale: 'en-US',
  timezoneId: 'America/New_York',
  viewport: { width: 1366, height: 768 },
};
if (storageStatePath && fs.existsSync(storageStatePath)) {
  contextOptions.storageState = storageStatePath;
}

const context = await browser.newContext(contextOptions);
const page = await context.newPage();
await context.addInitScript(() => {
  Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
});

if (sessionID) {
  await context.addCookies([{
    name: 'sessionid',
    value: sessionID,
    domain: '.tiktok.com',
    path: '/',
    httpOnly: true,
    secure: true,
    sameSite: 'None',
  }]);
}

const apiCandidates = [];
page.on('response', async (resp) => {
  try {
    const u = resp.url();
    if (!/tiktok\.com/.test(u)) return;
    if (!/(api|aweme|item_list|post|user\/detail|user\/post)/i.test(u)) return;
    const ct = resp.headers()['content-type'] || '';
    if (!/json|text|javascript/i.test(ct)) return;
    const txt = await resp.text();
    if (!txt || txt.length < 30) return;
    apiCandidates.push({ url: u, body: txt.slice(0, 4_000_000) });
  } catch (_) {
    // Best-effort diagnostics path.
  }
});

try {
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
  try {
    await page.waitForLoadState('networkidle', { timeout: 10000 });
  } catch (_) {
    // TikTok frequently keeps long-running connections.
  }
  try {
    await page.waitForSelector('a[href*="/video/"]', { timeout: 8000 });
  } catch (_) {
    // Continue to fallback parsing paths.
  }
  await page.waitForTimeout(4000);

  const profileApiMatches = [];
  const profileApiRegex = new RegExp(`(?:@${profile}|%40${profile}|\\\\/@${profile}|%2540${profile})`, 'i');
  const seenApi = new Set();
  for (const c of apiCandidates) {
    if (!profileApiRegex.test(c.url) && !profileApiRegex.test(c.body)) {
      continue;
    }
    const idMatches = c.body.matchAll(/"(?:aweme_id|id|item_id)"\s*:\s*"(\d{10,})"/g);
    for (const m of idMatches) {
      const id = m[1];
      if (seenApi.has(id)) continue;
      seenApi.add(id);
      let desc = '';
      const near = c.body.slice(Math.max(0, m.index - 1200), Math.min(c.body.length, m.index + 1200));
      const d = near.match(/"desc"\s*:\s*"([^"]{1,400})"/);
      if (d && d[1]) {
        desc = d[1]
          .replace(/\\u([0-9a-fA-F]{4})/g, (_, h) => String.fromCharCode(parseInt(h, 16)))
          .replace(/\\"/g, '"')
          .replace(/\\n/g, ' ')
          .trim();
      }
      profileApiMatches.push({
        id,
        url: `https://www.tiktok.com/@${profile}/video/${id}`,
        description: desc,
        download_url: '',
        created_at: new Date(0).toISOString(),
      });
      if (profileApiMatches.length >= limit * 3) break;
    }
    if (profileApiMatches.length >= limit * 3) break;
  }

  const result = await page.evaluate(({ max, profileName, apiVideos }) => {
    const primaryAnchors = Array.from(
      document.querySelectorAll('#user-post-item-list [data-e2e="user-post-item"] a[href*="/video/"]')
    );
    const anchors = primaryAnchors.length > 0
      ? primaryAnchors
      : Array.from(document.querySelectorAll('a[href*="/video/"]'));
    const out = [];
    const seen = new Set();

    for (const a of anchors) {
      const href = a.getAttribute('href') || '';
      const match = href.match(/\/video\/(\d+)/);
      if (!match) continue;
      const id = match[1];
      if (seen.has(id)) continue;
      const absolute = href.startsWith('http') ? href : `https://www.tiktok.com${href}`;

      const imgAlt = a.querySelector('img[alt]')?.getAttribute('alt') || '';
      out.push({
        id,
        url: absolute,
        description: imgAlt.trim(),
        download_url: '',
        created_at: new Date(0).toISOString(),
      });
      seen.add(id);
      if (out.length >= max) break;
    }

    const sourceByID = new Map();
    const addVideo = (id, desc = '', source = 'unknown') => {
      const sid = String(id || '');
      if (!sid || seen.has(sid)) return;
      out.push({
        id: sid,
        url: `https://www.tiktok.com/@${profileName}/video/${sid}`,
        description: (desc || '').trim(),
        download_url: '',
        created_at: new Date(0).toISOString(),
      });
      seen.add(sid);
      sourceByID.set(sid, source);
    };

    const extractProfileVideoIDs = (rawText) => {
      const ids = [];
      if (!rawText) return ids;
      const escapedProfile = profileName.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
      const patterns = [
        // https://www.tiktok.com/@gmbadass/video/123...
        new RegExp(`https?:(?:\\\\/|/){2}www\\\\.tiktok\\\\.com(?:\\\\/|/)@${escapedProfile}(?:\\\\/|/)video(?:\\\\/|/)(\\\\d{10,})`, 'g'),
        // /@gmbadass/video/123...
        new RegExp(`(?:\\\\/|/)@${escapedProfile}(?:\\\\/|/)video(?:\\\\/|/)(\\\\d{10,})`, 'g'),
      ];
      for (const re of patterns) {
        for (const m of rawText.matchAll(re)) {
          if (m[1]) ids.push(m[1]);
        }
      }
      return ids;
    };

    const parseState = (raw) => {
      if (!raw) return 0;
      let data;
      try {
        data = JSON.parse(raw);
      } catch (_) {
        return 0;
      }
      const before = out.length;
      const itemModule = data?.ItemModule || {};
      const lists = [
        data?.ItemList?.video?.list,
        data?.ItemList?.user?.post?.list,
        data?.UserModule?.users?.[Object.keys(data?.UserModule?.users || {})[0]]?.itemList,
      ];
      for (const list of lists) {
        if (!Array.isArray(list)) continue;
        for (const id of list) {
          const item = itemModule?.[id];
          addVideo(id, item?.desc || '', 'state-list');
          if (out.length >= max) break;
        }
        if (out.length >= max) break;
      }
      if (out.length < max) {
        for (const [id, item] of Object.entries(itemModule)) {
          addVideo(id, item?.desc || '', 'item-module');
          if (out.length >= max) break;
        }
      }

      // Generic deep walk fallback for unknown JSON structures.
      const visited = new WeakSet();
      const collectFromString = (s) => {
        if (typeof s !== 'string') return;
        const matches = s.matchAll(/(?:\/|\\\/)video(?:\/|\\\/)(\d{10,})/g);
        for (const m of matches) {
          addVideo(m[1], '');
          if (out.length >= max) break;
        }
      };
      const walk = (node) => {
        if (out.length >= max || node == null) return;
        if (typeof node === 'string') {
          collectFromString(node);
          return;
        }
        if (Array.isArray(node)) {
          for (const item of node) {
            walk(item);
            if (out.length >= max) break;
          }
          return;
        }
        if (typeof node !== 'object') return;
        if (visited.has(node)) return;
        visited.add(node);

        const idVal = node.id ?? node.aweme_id ?? node.itemId ?? node.video_id;
        const descVal = node.desc ?? node.description ?? node.title ?? '';
        if (idVal != null) {
          const sid = String(idVal);
          if (/^\d{10,}$/.test(sid)) {
            addVideo(sid, typeof descVal === 'string' ? descVal : '', 'deep-walk');
          }
        }

        for (const value of Object.values(node)) {
          walk(value);
          if (out.length >= max) break;
        }
      };
      if (out.length < max) {
        walk(data);
      }

      // Raw text fallback for escaped/non-JSON-typed fragments.
      if (out.length < max) {
        const profileIDs = extractProfileVideoIDs(raw);
        for (const id of profileIDs) {
          addVideo(id, '', 'profile-url-pattern');
          if (out.length >= max) break;
        }
      }
      return out.length - before;
    };

    const sigiRaw = document.querySelector('#SIGI_STATE')?.textContent || '';
    const rehydrationRaw = document.querySelector('#__UNIVERSAL_DATA_FOR_REHYDRATION__')?.textContent || '';
    const fromSigi = parseState(sigiRaw);
    const fromRehydration = parseState(rehydrationRaw);

    const profileIDsFromRaw = [
      ...extractProfileVideoIDs((document.querySelector('#SIGI_STATE')?.textContent || '') + '\n' + (document.querySelector('#__UNIVERSAL_DATA_FOR_REHYDRATION__')?.textContent || '')),
    ];

    let videos = out;
    if (Array.isArray(apiVideos) && apiVideos.length > 0) {
      const uniqueApi = [];
      const seen = new Set();
      for (const v of apiVideos) {
        if (!v || !v.id || seen.has(v.id)) continue;
        seen.add(v.id);
        uniqueApi.push(v);
      }
      videos = uniqueApi;
    }
    if (profileIDsFromRaw.length > 0) {
      const profileIDSet = new Set(profileIDsFromRaw);
      videos = videos.filter((v) => profileIDSet.has(v.id));
    }
    // Never trust deep-walk-only artifacts when nothing profile-specific was found.
    if (profileIDsFromRaw.length === 0) {
      videos = videos.filter((v) => (sourceByID.get(v.id) || '') !== 'deep-walk');
    }
    if (primaryAnchors.length === 0 || (Array.isArray(apiVideos) && apiVideos.length > 0)) {
      // Fallback-only mode: approximate "latest" by descending numeric ID.
      videos = [...videos].sort((a, b) => b.id.localeCompare(a.id, 'en'));
    }

    const antiBotHint =
      document.documentElement.outerHTML.includes('webmssdk') ||
      document.documentElement.outerHTML.includes('captcha') ||
      document.documentElement.outerHTML.includes('verify');
    const blockedLikely =
      antiBotHint &&
      primaryAnchors.length === 0 &&
      (apiVideos?.length || 0) === 0 &&
      profileIDsFromRaw.length === 0;

    return {
      videos: videos.slice(0, max),
      debug: {
        title: document.title,
        final_url: window.location.href,
        primary_anchor_count: primaryAnchors.length,
        anchor_count: anchors.length,
        sigi_bytes: sigiRaw.length,
        rehydration_bytes: rehydrationRaw.length,
        from_sigi: fromSigi,
        from_rehydration: fromRehydration,
        api_candidates: apiVideos?.length || 0,
        anti_bot_hint: antiBotHint,
        blocked_likely: blockedLikely,
        profile_ids_in_raw: profileIDsFromRaw.length,
        sample_profile_ids: profileIDsFromRaw.slice(0, 10),
        sample_api_ids: (apiVideos || []).slice(0, 10).map((v) => v.id),
        sample_selected_ids: videos.slice(0, 10).map((v) => v.id),
        sample_selected_sources: videos.slice(0, 10).map((v) => sourceByID.get(v.id) || ''),
      },
    };
  }, { max: limit, profileName: profile, apiVideos: profileApiMatches });

  const videos = result.videos || [];

  const title = await page.title();
  const finalURL = page.url();
  console.error(
    `[fetch_videos] open_url=${url} final_url=${finalURL} title=${JSON.stringify(title)} returned=${videos.length} primary_anchors=${result.debug?.primary_anchor_count ?? 0} anchors=${result.debug?.anchor_count ?? 0} sigi_bytes=${result.debug?.sigi_bytes ?? 0} rehydration_bytes=${result.debug?.rehydration_bytes ?? 0} from_sigi=${result.debug?.from_sigi ?? 0} from_rehydration=${result.debug?.from_rehydration ?? 0} api_candidates=${result.debug?.api_candidates ?? 0} profile_ids_in_raw=${result.debug?.profile_ids_in_raw ?? 0} anti_bot_hint=${result.debug?.anti_bot_hint ?? false} blocked_likely=${result.debug?.blocked_likely ?? false} storage_state=${storageStatePath ? (fs.existsSync(storageStatePath) ? 'loaded' : 'missing') : 'none'} sessionid=${sessionID ? 'set' : 'none'} stealth=on headless_mode=${headless ? 'new' : 'headed'}`
  );
  console.error(
    `[fetch_videos] sample_profile_ids=${JSON.stringify(result.debug?.sample_profile_ids ?? [])} sample_api_ids=${JSON.stringify(result.debug?.sample_api_ids ?? [])} sample_selected_ids=${JSON.stringify(result.debug?.sample_selected_ids ?? [])} sample_selected_sources=${JSON.stringify(result.debug?.sample_selected_sources ?? [])}`
  );
  if (videos.length === 0) {
    const html = await page.content();
    const snippet = html.slice(0, 800).replace(/\s+/g, ' ');
    console.error(`[fetch_videos] html_snippet=${JSON.stringify(snippet)}`);
  }

  for (const v of videos) {
    if (!v.download_url) {
      v.download_url = `https://www.tikwm.com/video/media/hdplay/${v.id}.mp4`;
    }
  }

  process.stdout.write(JSON.stringify(videos));
} finally {
  await browser.close();
}
