import { test } from 'node:test';
import assert from 'node:assert/strict';
import { NavController, type NavState } from '../src/nav/index.ts';

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
