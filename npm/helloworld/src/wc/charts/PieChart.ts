import * as echarts from 'echarts/core';
import { PieChart as ePieChart } from 'echarts/charts';
import { CanvasRenderer } from 'echarts/renderers';
import { LitElement, html, css, nothing, PropertyValues } from 'lit';
import { customElement, property } from 'lit/decorators.js';

// Register the required components
echarts.use([
  ePieChart,
  CanvasRenderer
]);

/**
 * @class PieChart
 *
 * This class is a renders a pie chart, within a canvas or content element.
 *
 * @example
 * <wc-canvas vertical>
 *  <wc-nav>....</wc-nav>
 *  <wc-piechart>....</wc-piechart>
 * </wc-canvas>
 */
@customElement('wc-piechart')
export class PieChart extends LitElement {
  #chart: echarts.ECharts | null = null;
  #node: HTMLDivElement | null = null;

  protected firstUpdated(): void {    
    this.#chart = echarts.init(this.shadowRoot.querySelector('div'));

    // Set options for the chart
    this.#chart.setOption({
      series: [
        {
          type: 'pie',
          label: {
            "fontFamily": "IBM Plex Sans",
          },
          data: [
            {
              value: 335,
              name: 'Direct Visit'
            },
            {
              value: 234,
              name: 'Union Ad'
            },
            {
              value: 1548,
              name: 'Search Engine'
            }
          ]
        }
      ]
    });

    // Set to resize the chart when the window is resized
    window.addEventListener('resize', function () {
      this.#chart.resize();
    }.bind(this));
  }


  static get styles() {
    return css`
        :host {
          flex: 1 0;
      }

      div {
        display: flex;
        flex: 1 0;
        height: 100%;
      }
    `;
  }

  render() {
    return html`
      <div class=${this.className || nothing}>${this.#node}</div>
    `;
  }

  get className() {
    const classes = [];
    return classes.join(' ');
  }
}
