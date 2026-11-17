import { test } from 'node:test';
import assert from 'node:assert/strict';
import { NavController, navItems, type NavState } from '../src/nav/index.ts';

test('NavController', () => {
    // The pages are grouped by their sections, in order, and the active
    // page's section is active, and links to its first page
    const nav = new NavController({
        name: 'manager',
        prefix: 'pg',
        href: '#',
        pages: [
            { id: 'databases', label: 'Databases', href: '#databases', section: 'Data' },
            { id: 'roles', label: 'Roles', href: '#roles', section: 'Access' },
            { id: 'schemas', label: 'Schemas', href: '#schemas', section: 'Data' },
            { id: 'about', label: 'About', href: '#about' },
        ],
    });
    const states: NavState[] = [];
    nav.subscribe((state) => states.push(state));
    nav.setActive('schemas');
    nav.setActive('schemas');
    assert.deepEqual(nav.sections.map(({ label, href, active, pages }) => ({ label, href, active, pages: pages.map((page) => page.id) })), [
        { label: 'Data', href: '#databases', active: true, pages: ['databases', 'schemas'] },
        { label: 'Access', href: '#roles', active: false, pages: ['roles'] },
        { label: 'About', href: '#about', active: false, pages: ['about'] },
    ]);
    assert.deepEqual(states.map((state) => state.active), [undefined, 'schemas']);

    // The pages can be changed
    nav.setPages([{ id: 'roles', label: 'Roles', href: '#roles' }]);
    assert.deepEqual(nav.sections.map((section) => section.label), ['Roles']);
});

test('NavController sections', async (t) => {
    const nav = new NavController({
        name: 'app',
        href: '#',
        pages: [
            { id: 'credentials', label: 'Credentials', href: '#credentials', section: 'Credentials' },
            { id: 'queues', label: 'Queues', href: '#queues', section: 'Queues' },
            { id: 'tasks', label: 'Tasks', href: '#tasks', section: 'Queues' },
        ],
    });

    await t.test('the active section is the first, until a page is active', () => {
        assert.equal(nav.activeSection?.label, 'Credentials');
        nav.setActive('tasks');
        assert.equal(nav.activeSection?.label, 'Queues');
        assert.deepEqual(nav.activeSection?.pages.map((page) => page.id), ['queues', 'tasks']);
    });

    await t.test('a module replaces its own section, which stays where it is', () => {
        const states: NavState[] = [];
        const unsubscribe = nav.subscribe((state) => states.push(state));
        nav.setSection('Credentials', [
            { id: 'credentials', label: 'All', href: '#credentials' },
            { id: 'rotate', label: 'Needs rotation', href: '#rotate' },
        ]);
        unsubscribe();
        assert.deepEqual(nav.sections.map((section) => [section.label, section.pages.map((page) => page.id)]), [
            ['Credentials', ['credentials', 'rotate']],
            ['Queues', ['queues', 'tasks']],
        ]);
        assert.equal(nav.current.pages[1].section, 'Credentials', 'the pages are given the section');
        assert.equal(states.length, 2, 'subscribers are told');
        assert.equal(nav.current.active, 'tasks', 'the active page is kept');
    });

    await t.test('a new section is added at the end, and an empty one removed', () => {
        nav.setSection('Users', [{ id: 'users', label: 'Users', href: '#users' }]);
        assert.deepEqual(nav.sections.map((section) => section.label), ['Credentials', 'Queues', 'Users']);
        nav.setSection('Queues', []);
        assert.deepEqual(nav.sections.map((section) => section.label), ['Credentials', 'Users']);
    });
});

test('NavController pages without a section', () => {
    // A page without a section is a section of its own, even when its label
    // is a named section's, or another page's
    const nav = new NavController({
        name: 'app',
        href: '#',
        pages: [
            { id: 'credentials', label: 'Credentials', href: '#credentials', section: 'Credentials' },
            { id: 'about', label: 'Credentials', href: '#about' },
            { id: 'help', label: 'Help', href: '#help' },
            { id: 'help2', label: 'Help', href: '#help2' },
        ],
    });
    assert.deepEqual(nav.sections.map((section) => [section.label, section.href, section.pages.map((page) => page.id)]), [
        ['Credentials', '#credentials', ['credentials']],
        ['Credentials', '#about', ['about']],
        ['Help', '#help', ['help']],
        ['Help', '#help2', ['help2']],
    ]);

    // Replacing a section leaves the page without a section of its label
    nav.setSection('Credentials', [{ id: 'all', label: 'All', href: '#all' }]);
    assert.deepEqual(nav.current.pages.map((page) => page.id), ['all', 'about', 'help', 'help2']);
    nav.setActive('about');
    assert.equal(nav.activeSection?.href, '#about');
});

test('NavController modules', () => {
    // Modules share a section, each replacing only its own pages, which stay
    // where they are, and a new module's are after the section's
    const nav = new NavController({ name: 'llm', href: '#', pages: [] });
    nav.setModule('chat', 'llm', [{ id: 'chat', label: 'Conversations', href: '#chat' }]);
    nav.setModule('models', 'llm', [{ id: 'models', label: 'Models', href: '#models' }]);
    nav.setModule('other', 'Other', [{ id: 'other', label: 'Other', href: '#other' }]);
    nav.setModule('tools', 'llm', [{ id: 'tools', label: 'Tools', href: '#tools' }]);
    assert.deepEqual(nav.current.pages.map((page) => page.id), ['chat', 'models', 'tools', 'other']);

    nav.setModule('models', 'llm', [
        { id: 'models', label: 'Models', href: '#models' },
        { id: 'providers', label: 'Providers', href: '#providers' },
    ]);
    assert.deepEqual(nav.current.pages.map((page) => page.id), ['chat', 'models', 'providers', 'tools', 'other']);
    assert.deepEqual(nav.sections.map((section) => section.label), ['llm', 'Other']);
    assert.equal(nav.current.pages[1].module, 'models');
    assert.equal(nav.current.pages[1].section, 'llm');

    nav.setModule('models', 'llm', []);
    assert.deepEqual(nav.current.pages.map((page) => page.id), ['chat', 'tools', 'other']);
});

test('navItems', () => {
    // Pages next to each other with the same menu are in it, and others are
    // on their own
    const items = navItems([
        { id: 'chat', label: 'Conversations', href: '#chat' },
        { id: 'notes', label: 'Notes', href: '#workspace/notes', menu: 'Workspaces' },
        { id: 'files', label: 'Files', href: '#workspace/files', menu: 'Workspaces' },
        { id: 'tools', label: 'Tools', href: '#tools' },
    ]);
    assert.deepEqual(items.map((item) => ({ menu: item.menu, pages: item.pages.map((page) => page.id) })), [
        { menu: undefined, pages: ['chat'] },
        { menu: 'Workspaces', pages: ['notes', 'files'] },
        { menu: undefined, pages: ['tools'] },
    ]);
});
