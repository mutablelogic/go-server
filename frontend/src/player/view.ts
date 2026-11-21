// The gs-media-player element plays audio, and video, such as in the
// sidebar, above gs-server-status, so it's played in the page, rather than
// in a new tab, and keeps playing as the page changes. It plays the media of
// each PlayEvent on the document, or of play, from its start, in place of
// what it plays, and is hidden until something is played, and when it's
// closed. It says what it plays is no longer available, when it can't be
// loaded. Its playing attribute is set while it plays, so an app can make
// room for it, such as:
//
//     cds-side-nav:has(gs-media-player[playing]) cds-side-nav-items {
//         margin-bottom: calc(8rem + 2px);
//     }

import { LitElement, css, html, nothing, type PropertyValues } from 'lit';
import { keyed } from 'lit/directives/keyed.js';
import { closeIcon, musicIcon, videoIcon } from '../internal/icons.ts';
import { PlayEvent, type Media } from './index.ts';

// Carbon's components, which the view renders
import '@carbon/web-components/es/components/icon-button/index.js';

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

export const MediaPlayerTag = 'gs-media-player';

// MediaPlayerClosed is the event of the player, once it's closed, or stops
export const MediaPlayerClosed = 'gs-media-player-closed';

///////////////////////////////////////////////////////////////////////////////
// ELEMENT

export class MediaPlayerElement extends LitElement {
    static override properties = {
        media: { attribute: false },
        playing: { type: Boolean, reflect: true },
        unavailable: { state: true },
        requests: { state: true },
    };

    // media is what's played, or none, when it's hidden, playing is whether
    // something is, which is reflected, and unavailable is whether what's
    // played can't be loaded, such as when it was deleted
    declare media?: Media;
    declare playing: boolean;
    declare private unavailable: boolean;

    // requests counts what's asked to be played, so each is played, from its
    // start, of an element of its own, even of the same media, or URL
    declare private requests: number;

    private readonly onPlay = (event: Event) => this.play((event as CustomEvent<Media>).detail);

    static override styles = css`
        :host {
            position: absolute;
            right: 0;
            bottom: calc(3rem + 1px);
            left: 0;
            display: none;
            padding: 0.5rem 0.5rem 0.75rem 1rem;
            border-top: 1px solid var(--cds-border-subtle);
            color: var(--cds-text-secondary);
            font-size: 0.75rem;
            letter-spacing: 0.32px;
        }

        :host([playing]) {
            display: block;
        }

        .title {
            display: flex;
            align-items: center;
            gap: 0.375rem;
            margin-bottom: 0.25rem;
        }

        .title svg {
            flex: none;
            fill: currentColor;
        }

        .name {
            flex: 1;
            min-width: 0;
            overflow: hidden;
            color: var(--cds-text-primary);
            text-overflow: ellipsis;
            white-space: nowrap;
        }

        audio,
        video {
            display: block;
            width: calc(100% - 0.5rem);
            color-scheme: dark;
        }

        video {
            max-height: 10rem;
            background: black;
        }

        .unavailable {
            margin: 0.25rem 0;
            font-style: italic;
        }
    `;

    constructor() {
        super();
        this.playing = false;
        this.unavailable = false;
        this.requests = 0;
    }

    override connectedCallback() {
        super.connectedCallback();
        document.addEventListener(PlayEvent, this.onPlay);
    }

    override disconnectedCallback() {
        super.disconnectedCallback();
        document.removeEventListener(PlayEvent, this.onPlay);
        this.close();
    }

    // play plays the media, from its start, in place of what's played, even
    // when it's what's played
    play(media: Media) {
        this.unload();
        this.unavailable = false;
        this.media = media;
        this.requests++;
    }

    // close stops playing, and closes the media's request, as the browser
    // can keep it open while the element is paused, or until it's collected,
    // such as when the server is stopping, and hides the player
    close() {
        this.unload();
        const closed = this.media !== undefined;
        this.media = undefined;
        this.unavailable = false;
        if (closed) {
            this.dispatchEvent(new CustomEvent(MediaPlayerClosed, { bubbles: true, composed: true }));
        }
    }

    protected override willUpdate(changed: PropertyValues<this>) {
        if (changed.has('media')) {
            this.playing = this.media !== undefined;
        }
    }

    // The media which is chosen is played, once its element is rendered
    protected override updated(changed: PropertyValues<this>) {
        if ((changed as Map<string, unknown>).has('requests') && this.media) {
            this.element?.play().catch(() => {
                // The browser didn't let it play, so its controls play it
            });
        }
    }

    protected override render() {
        const media = this.media;
        if (!media) {
            return nothing;
        }
        const video = (media.type ?? '').toLowerCase().startsWith('video/');
        return html`
            <div class="title">
                ${video ? videoIcon : musicIcon}
                <span class="name" title=${media.name}>${media.name}</span>
                <cds-icon-button kind="ghost" size="sm" align="top-right" @click=${() => this.close()}>
                    ${closeIcon}
                    <span slot="tooltip-content">Close</span>
                </cds-icon-button>
            </div>
            ${this.unavailable
                ? html`<div class="unavailable">This file is no longer available</div>`
                : keyed(this.requests, video
                    ? html`<video controls preload="metadata" src=${media.url} @error=${() => this.failed()}></video>`
                    : html`<audio controls preload="metadata" src=${media.url} @error=${() => this.failed()}></audio>`)}
        `;
    }

    // unload stops the element which plays, and closes its request, before
    // it's replaced, or removed
    private unload() {
        const element = this.element;
        if (element) {
            element.pause();
            element.removeAttribute('src');
            element.load();
        }
    }

    // element is the audio, or video, which plays the media
    private get element(): HTMLMediaElement | null | undefined {
        return this.shadowRoot?.querySelector('audio, video');
    }

    // failed says what's played is no longer available, when it can't be
    // loaded, rather than leaving a player which can't play
    private failed() {
        this.element?.pause();
        this.unavailable = true;
    }
}

if (!customElements.get(MediaPlayerTag)) {
    customElements.define(MediaPlayerTag, MediaPlayerElement);
}

declare global {
    interface HTMLElementTagNameMap {
        [MediaPlayerTag]: MediaPlayerElement;
    }
}
