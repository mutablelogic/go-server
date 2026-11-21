// What a player plays, which is audio, or video, and how a page plays it: it
// dispatches PlayEvent, of the media, which the gs-media-player element, of
// @mutablelogic/go-server/player/view, plays, wherever it is in the page

///////////////////////////////////////////////////////////////////////////////
// TYPES

// Media is what's played: its URL, its name, and its type, such as
// audio/mpeg, or video/mp4
export interface Media {
    url: string;
    name: string;
    type?: string;
}

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

// PlayEvent is the event which plays media, whose detail is the media,
// which the player listens for on the document
export const PlayEvent = 'gs-media-play';

///////////////////////////////////////////////////////////////////////////////
// PUBLIC FUNCTIONS

// playable returns whether media of the type is played, which is audio, and
// video
export function playable(type?: string): boolean {
    const mediaType = (type ?? '').split(';')[0].trim().toLowerCase();
    return mediaType.startsWith('audio/') || mediaType.startsWith('video/');
}

// play plays the media, in the page's player, of the target, which is the
// document, by default
export function play(media: Media, target: EventTarget = document): void {
    target.dispatchEvent(new CustomEvent<Media>(PlayEvent, { detail: media, bubbles: true, composed: true }));
}
