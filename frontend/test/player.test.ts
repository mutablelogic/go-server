import { test } from 'node:test';
import assert from 'node:assert/strict';
import { play, playable, PlayEvent, type Media } from '../src/player/index.ts';

test('player', async (t) => {
    await t.test('playable', () => {
        // Audio and video are played, of their types, with parameters and in
        // any case, and nothing else is
        for (const type of ['audio/mpeg', 'video/mp4', 'Audio/WAV', 'video/webm; codecs="vp9"']) {
            assert.equal(playable(type), true, type);
        }
        for (const type of [undefined, '', 'image/png', 'application/pdf', 'text/plain']) {
            assert.equal(playable(type), false, String(type));
        }
    });

    await t.test('play', () => {
        // Playing media dispatches its event, of the media, on the target
        const target = new EventTarget();
        const played: Media[] = [];
        target.addEventListener(PlayEvent, (event) => played.push((event as CustomEvent<Media>).detail));
        const media = { url: '/api/object/music/a.mp3', name: 'a.mp3', type: 'audio/mpeg' };
        play(media, target);
        assert.deepEqual(played, [media]);
    });
});
