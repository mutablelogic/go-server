import { LitElement, html, css, nothing } from 'lit';

/**
 * @class CanvasElement
 *
 * This class is used to contain  content boxes which are stacked
 * vertically or horizontally within the canvas.
 *
 * @property {Boolean} vertical - Fill the canvas vertically rather than horizontally, default false
 * @property {String} theme - Canvas design theme - light dark
 *
 * @example
 * <c-canvas vertical>
 *  <c-canvas-content>....</c-canvas-content>
 *  <c-canvas-content>....</c-canvas-content>
 *  <c-canvas-content>....</c-canvas-content>
 * </c-canvas>
 */
export class CanvasElement extends LitElement {
  static get localName() {
    return 'c-canvas';
  }

  constructor() {
    super();
    this.theme = 'light';
    this.vertical = false;
  }

  static get properties() {
    return {
      vertical: { type: Boolean },
      theme: { type: String },
    };
  }

  static get styles() {
    return css`
      div {
        position: absolute;
        top: 0;
        bottom: 0;
        left: 0;
        right: 0;
        display: flex;
      }

      /* Any canvas section is in flexbox mode with a fixed size */
      ::slotted(c-canvas-content), ::slotted(c-canvas-nav) {
        border: 0.1px solid red !important;
      }

      div.vertical {
        flex-direction: column;

        /* The first section in vertical mode gets a border below */
        > ::slotted(c-canvas-nav:first-child), ::slotted(c-canvas-content:first-child) {
          min-height: 40px;
          border-bottom-style: solid;
        }

        /* The last section in vertical mode gets a border above */
        > ::slotted(c-canvas-nav:last-child), ::slotted(c-canvas-content:last-child) {
            min-height: 30px;
            border-top-style: solid;
        }
      }

      div:not(.vertical) {
        flex-direction: row;

        /* The first section in horizontal mode gets a border right */
        > ::slotted(c-canvas-nav:first-child), ::slotted(c-canvas-content:first-child) {
          min-width: 60px;
          border-right-style: solid;
        }

        /* The last section in horizonal mode gets a border left */
        > ::slotted(c-canvas-nav:last-child), ::slotted(c-canvas-content:last-child) {
          min-width: 60px;
          border-left-style: solid;
        }
      }

      /* Flex containers stretch */
      ::slotted(c-canvas-content[flex]) {
        flex: 999 0;
        overflow: auto;        
      }

      /* Hidden navbars are hidden */
      ::slotted(c-canvas-nav[hidden]) {
        display: none;
      }

      /* Light theme setting colours and border widths */
      div.theme-light {
        & ::slotted(c-canvas-nav), ::slotted(c-canvas-content) {
          background-color: var(--light-color);
          color: var(--dark-color);
          border-color: var(--grey-20-color);
          border-width: 1px;
        }
      }

      /* Dark theme setting colours and border widths */
      div.theme-dark {
        & ::slotted(c-canvas-nav), ::slotted(c-canvas-content) {
          background-color: var(--dark-color);
          color: var(--light-color);
          border-color: var(--grey-40-color);
          border-width: 1px;
        }
      }
    `;
  }

  render() {
    return html`
      <div class=${this.className || nothing}>
        <slot></slot>
      </div>
    `;
  }

  get className() {
    const classes = [];
    if (this.theme) {
      classes.push(`theme-${this.theme}`);
    }
    if (this.vertical) {
      classes.push('vertical');
    }
    return classes.join(' ');
  }
}
