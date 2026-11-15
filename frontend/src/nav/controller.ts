// The navigation controller keeps an app's name, and its pages, in sections,
// and which is active, which the router sets, and tells the views which
// subscribe to it, such as gs-header, whose menu bar is the sections, and
// gs-side-nav, whose sidebar is the active section's pages, whenever it
// changes. An app can be composed of modules, which each own a section, and
// its pages, which setSection replaces.

// NavPage is a page of the app: its id, which the router routes to, its
// label, its link, and its section, such as Databases
export interface NavPage {
    id: string;
    label: string;
    href: string;
    section?: string;
}

// NavState is the app's name, and its prefix, such as pg and manager, its
// pages, and the active page's id
export interface NavState {
    name: string;
    prefix?: string;
    href: string;
    pages: readonly NavPage[];
    active?: string;
}

// NavSection is a section of the pages, which the menu bar shows, and the
// sidebar groups the pages by
export interface NavSection {
    label: string;
    href: string;
    pages: readonly NavPage[];
    active: boolean;
}

export class NavController {
    private readonly listeners = new Set<(state: NavState) => void>();
    private state: NavState;

    constructor(state: Omit<NavState, 'active'>) {
        this.state = { ...state, pages: [...state.pages] };
    }

    // current is the navigation's state
    get current(): NavState {
        return this.state;
    }

    // sections returns the pages, by their sections, in the order of their
    // first pages, with whether each has the active page, and its link,
    // which is its first page's. Pages without a section are a section of
    // their own, with their label.
    get sections(): NavSection[] {
        const sections: NavSection[] = [];
        for (const page of this.state.pages) {
            const label = page.section ?? page.label;
            let section = sections.find((section) => section.label === label);
            if (!section) {
                section = { label, href: page.href, pages: [], active: false };
                sections.push(section);
            }
            section.pages = [...section.pages, page];
            section.active ||= page.id === this.state.active;
        }
        return sections;
    }

    // activeSection is the section of the active page, or the first section,
    // when no page is active, whose pages the sidebar shows
    get activeSection(): NavSection | undefined {
        const sections = this.sections;
        return sections.find((section) => section.active) ?? sections[0];
    }

    // setActive makes the page with the id active, such as when it's routed
    // to, or none
    setActive(id?: string) {
        if (id !== this.state.active) {
            this.update({ active: id });
        }
    }

    // setPages sets the app's pages, such as once it knows what it can show
    setPages(pages: NavPage[]) {
        this.update({ pages: [...pages] });
    }

    // setSection replaces the pages of the section, such as a module's, which
    // owns it, keeping the other sections' pages, and where the section is in
    // the menu, or adding it at the end. The pages are given the section.
    setSection(section: string, pages: NavPage[]) {
        const owned = pages.map((page) => ({ ...page, section }));
        const result: NavPage[] = [];
        let placed = false;
        for (const page of this.state.pages) {
            if ((page.section ?? page.label) !== section) {
                result.push(page);
            } else if (!placed) {
                result.push(...owned);
                placed = true;
            }
        }
        if (!placed) {
            result.push(...owned);
        }
        this.update({ pages: result });
    }

    // subscribe calls the function with the state, now, and whenever it
    // changes, until the function it returns is called, which unsubscribes
    subscribe(fn: (state: NavState) => void): () => void {
        this.listeners.add(fn);
        fn(this.state);
        return () => this.listeners.delete(fn);
    }

    private update(state: Partial<NavState>) {
        this.state = { ...this.state, ...state };
        this.listeners.forEach((fn) => fn(this.state));
    }
}
