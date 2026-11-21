// renderIcon renders an action's icon: a Carbon icon descriptor, such as
// @carbon/icons/es/add/16.js's default export, in the slot, when it's given,
// or else an icon which is rendered already, as it is

import { html } from 'lit';
import iconLoader from '@carbon/web-components/es/globals/internal/icon-loader.js';
import type { CarbonIconDescriptor } from '@carbon/web-components/es/globals/internal/icon-loader-utils.js';

export function renderIcon(icon: unknown, slot?: string) {
    if (isDescriptor(icon)) {
        return iconLoader(icon, slot ? { slot } : undefined);
    }
    return slot ? html`<span slot=${slot}>${icon}</span>` : icon;
}

// isDescriptor is true for a Carbon icon descriptor
function isDescriptor(icon: unknown): icon is CarbonIconDescriptor {
    return typeof icon === 'object' && icon !== null && 'elem' in icon && 'attrs' in icon;
}
