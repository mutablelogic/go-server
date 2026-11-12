// The router shows the page of the location's hash, such as #roles, and
// hides the others, as the location changes, which every app of a server
// has

// Route is a page, and its parameters, from the location's hash, whose
// elements are separated by /, such as #role/admin
export interface Route {
    readonly page: string;
    readonly params: readonly string[];
}

// parseRoute returns the route of a location's hash, or the default page,
// when it's empty
export function parseRoute(hash: string, defaultPage: string): Route {
    const [page, ...params] = hash.replace(/^#/, '').split('?')[0].split('/').map((element) => {
        try {
            return decodeURIComponent(element);
        } catch {
            return element;
        }
    });
    return { page: page || defaultPage, params };
}

export class Router {
    private readonly pages = new Map<string, { element: HTMLElement; enter?: (route: Route) => void }>();
    private readonly listeners: ((route: Route) => void)[] = [];

    private readonly defaultPage: string;

    constructor(defaultPage: string) {
        this.defaultPage = defaultPage;
    }

    // add adds a page, which is shown for routes to its name, calling enter,
    // when it's set, with the route
    add(name: string, element: HTMLElement | null, enter?: (route: Route) => void): this {
        if (element) {
            this.pages.set(name, { element, enter });
        }
        return this;
    }

    // onRoute calls the function with every route
    onRoute(fn: (route: Route) => void): this {
        this.listeners.push(fn);
        return this;
    }

    // start shows the page of the location, and follows the location
    start() {
        window.addEventListener('hashchange', () => this.show());
        this.show();
    }

    private show() {
        const route = parseRoute(location.hash, this.defaultPage);
        const shown = this.pages.get(route.page) ?? this.pages.get(this.defaultPage);
        for (const page of this.pages.values()) {
            page.element.hidden = page !== shown;
        }
        shown?.enter?.(route);
        this.listeners.forEach((fn) => fn(route));
    }
}
