// The navigation controller keeps an app's name, and its pages, in sections,
// and which is active, which the router sets, and tells the views which
// subscribe to it, such as gs-header, whose menu bar is the sections, and
// gs-side-nav, whose sidebar is the active section's pages, whenever it
// changes. An app can be composed of modules, which each own a section, and
// its pages, which setSection replaces.

// NavPage is a page of the app: its id, which the router routes to, its
// label, its link, its section, such as Databases, the module which owns
// it, when modules share a section, and the menu it's in, in the sidebar,
// such as Workspaces, which the pages next to it with the same menu are in,
// or the id of the page it's beneath, in the sidebar, such as a volume's,
// which is beneath Volumes, indented, and isn't collapsed, and its title,
// which the sidebar shows when it's pointed at, such as its whole label,
// which the sidebar cuts short, when it's long
export interface NavPage {
    id: string;
    label: string;
    title?: string;
    href: string;
    section?: string;
    module?: string;
    menu?: string;
    parent?: string;
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

// NavItem is an item of the sidebar: a page, on its own, without a menu,
// with the pages beneath it, which are indented, or the pages of a menu
export interface NavItem {
    menu?: string;
    pages: NavPage[];
    children: NavPage[];
}

// navItems returns the sidebar's items of the pages: each page on its own,
// with the pages whose parent it is beneath it, in their order, or, with the
// pages right before it with the same menu, in it. A page with a parent is
// never in a menu: one whose parent isn't a page on its own, of the pages,
// is on its own.
export function navItems(pages: readonly NavPage[]): NavItem[] {
    const parents = new Set(pages.filter((page) => page.menu === undefined && page.parent === undefined).map((page) => page.id));
    const isChild = (page: NavPage) => page.parent !== undefined && page.parent !== page.id && parents.has(page.parent);
    const result: NavItem[] = [];
    const byParent = new Map<string, NavItem>();

    // previous is the item of the page before, so a menu only has the
    // pages next to each other, and not those which a child was between
    let previous: NavItem | undefined;
    for (const page of pages) {
        if (isChild(page)) {
            previous = undefined;
            continue;
        }
        const menu = page.parent === undefined ? page.menu : undefined;
        if (menu !== undefined && previous?.menu === menu) {
            previous.pages.push(page);
            continue;
        }
        const item: NavItem = { menu, pages: [page], children: [] };
        result.push(item);
        previous = item;
        if (menu === undefined && page.parent === undefined) {
            byParent.set(page.id, item);
        }
    }
    for (const page of pages) {
        if (isChild(page)) {
            byParent.get(page.parent!)?.children.push(page);
        }
    }
    return result;
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
        // Pages are grouped by their sections, and a page without one is a
        // section of its own, even when its label is another's
        const sections: NavSection[] = [];
        const named = new Map<string, NavSection>();
        for (const page of this.state.pages) {
            let section = page.section === undefined ? undefined : named.get(page.section);
            if (!section) {
                section = { label: page.section ?? page.label, href: page.href, pages: [], active: false };
                sections.push(section);
                if (page.section !== undefined) {
                    named.set(page.section, section);
                }
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
    // the menu, or adding it at the end. The pages are given the section, and
    // pages without a section are never replaced, even with its label.
    setSection(section: string, pages: NavPage[]) {
        const owned = pages.map((page) => ({ ...page, section }));
        const result: NavPage[] = [];
        let placed = false;
        for (const page of this.state.pages) {
            if (page.section !== section) {
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

    // setModule replaces the pages of the module, in the section, which
    // other modules can share, such as an app whose modules' pages are all in
    // one sidebar, keeping the other pages, and where the module's pages
    // are, or adding them at the end of the section, or of the pages, when
    // it's a new section. The pages are given the module, and the section.
    setModule(module: string, section: string, pages: NavPage[]) {
        const owned = pages.map((page) => ({ ...page, section, module }));
        const result: NavPage[] = [];
        let placed = false;
        for (const page of this.state.pages) {
            if (page.module !== module) {
                result.push(page);
            } else if (!placed) {
                result.push(...owned);
                placed = true;
            }
        }
        if (!placed) {
            // After the section's last page, or at the end
            let last = -1;
            result.forEach((page, i) => {
                if (page.section === section) {
                    last = i;
                }
            });
            result.splice(last < 0 ? result.length : last + 1, 0, ...owned);
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
