import { LitElement, html, css, nothing } from 'lit';
import { customElement, property } from 'lit/decorators.js';

/**
 * @class DataTable
 *
 * This class is a renders a table of data, within a content element.
 *
 * @example
 * <wc-datatable>
 * </wc-datatable>
 */
@customElement('wc-datatable')
export class DataTable extends LitElement {
  @property({ type: Boolean }) striped: boolean;
  @property({ type: Boolean }) selectable: boolean;

  static get styles() {
    return css`
      :host {
          flex: 1 0;
      }
      table {
        border: var(--table-border);
        border-radius: var(--table-border-radius);
        inline-size: 100%;
        border-collapse: collapse;
        border-spacing: 0;
      }
      thead {
        background-color: var(--table-background-color-head);
        color: var(--table-color-head);
        block-size: 3rem;
      }
      thead th {
        text-align: start;
        padding: var(--table-padding-head);
        font-size: var(--table-font-size-head);
        font-weight: var(--table-font-weight-head);
        user-select: none;
      }
      tbody {
        background-color: var(--table-background-color);
        color: var(--table-color);
        block-size: 3rem;
      }
      table.striped tbody tr:nth-child(even) {
        background-color: var(--table-background-color-striped);
      }
      tbody tr {
        border-bottom: var(--table-border-bottom-row);
      }
      table.selectable tbody tr {
        cursor: pointer;
      } 
      table.selectable tbody tr:hover {
        background-color: var(--table-background-color-hover);
      }
      tbody tr {
        font-size: var(--table-font-size-row);
        font-weight: var(--table-font-weight-row);
        border-bottom: var(--table-border-bottom-row);
      }
      tbody td {
        text-align: start;
        padding: var(--table-padding-cell);
      }
    `;
  }

  render() {
    return html`
      <table class=${this.className || nothing}>
        <thead>
          <tr>
            <th>Month</th>
            <th>Value</th>
            <th>Total</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Jan</td>
            <td>&pound;100.20</td>
            <td>&pound;100.20</td>
          </tr>
          <tr>
            <td>Feb</td>
            <td>&pound;120.50</td>
            <td>&pound;220.70</td>
          </tr>
          <tr>
            <td>Mar</td>
            <td>&pound;90.75</td>
            <td>&pound;311.45</td>
          </tr>
          <tr>
            <td>Apr</td>
            <td>&pound;150.00</td>
            <td>&pound;461.45</td>
          </tr>
          <tr>
            <td>May</td>
            <td>&pound;110.30</td>
            <td>&pound;571.75</td>
          </tr>
          <tr>
            <td>Jun</td>
            <td>&pound;130.80</td>
            <td>&pound;702.55</td>
          </tr>
          <tr>
            <td>Jul</td>
            <td>&pound;140.15</td>
            <td>&pound;842.70</td>
          </tr>
          <tr>
            <td>Aug</td>
            <td>&pound;105.60</td>
            <td>&pound;948.30</td>
          </tr>
          <tr>
            <td>Sep</td>
            <td>&pound;160.90</td>
            <td>&pound;1109.20</td>
          </tr>
          <tr>
            <td>Oct</td>
            <td>&pound;125.40</td>
            <td>&pound;1234.60</td>
          </tr>
          <tr>
            <td>Nov</td>
            <td>&pound;115.25</td>
            <td>&pound;1349.85</td>
          </tr>
          <tr>
            <td>Dec</td>
            <td>&pound;180.55</td>
            <td>&pound;1530.40</td>
          </tr>
        </tbody>
      </table>
    `;
  }

  get className() {
    const classes = [];
    if (this.striped) {
      classes.push('striped');
    }
    if (this.selectable) {
      classes.push('selectable');
    }
    return classes.join(' ');
  }
}
