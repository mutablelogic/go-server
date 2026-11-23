import { test } from 'node:test';
import assert from 'node:assert/strict';
import { PanelController, type PanelAction, type PanelState, type PanelTab } from '../src/panel/index.ts';

test('PanelController', async (t) => {
    await t.test('open', () => {
        // The panel shows its item while it's open, and keeps it while it
        // closes, and a subscriber is told of its state now, and whenever it
        // changes, and the close functions are called with the item
        const panel = new PanelController<string>({ header: { title: 'Thing' } });
        const states: PanelState<string>[] = [];
        panel.subscribe((state) => states.push(state));
        const closed: string[] = [];
        panel.onClose((item) => closed.push(item));
        panel.open('one');
        assert.ok(panel.isOpen);
        assert.equal(panel.item, 'one');
        panel.close();
        panel.close();
        assert.ok(!panel.isOpen);
        assert.equal(panel.item, 'one');
        assert.deepEqual(closed, ['one']);
        assert.deepEqual(states.map((state) => state.open), [false, true, false]);
        assert.equal(states[0].header.title, 'Thing');
        assert.ok(states[0].closable);
    });

    await t.test('header', () => {
        const panel = new PanelController({ closable: false });
        panel.setHeader({ title: 'Edit', subtitle: 'An agent', icon: '/logo.svg' });
        assert.deepEqual(panel.current.header, { title: 'Edit', subtitle: 'An agent', icon: '/logo.svg' });
        assert.ok(!panel.current.closable);
        panel.setClosable(true);
        assert.ok(panel.current.closable);
    });

    await t.test('save', async () => {
        // An action runs until what it returns settles, while which it, and
        // the others, can't run, and then closes the panel
        let resolve!: () => void;
        const save: PanelAction = { label: 'Save', run: () => new Promise<void>((done) => { resolve = done; }) };
        const cancel: PanelAction = { label: 'Cancel', kind: 'secondary' };
        const panel = new PanelController<string>({ actions: [save, cancel] });
        panel.open('one');
        const saving = panel.run(save);
        assert.ok(panel.running);
        assert.deepEqual(panel.current.actions.map((action) => action.running), [true, false]);
        assert.equal(await panel.run(cancel), false);
        assert.equal(await panel.run(save), false);
        assert.ok(panel.isOpen);
        resolve();
        assert.equal(await saving, true);
        assert.ok(!panel.running);
        assert.ok(!panel.isOpen);
    });

    await t.test('cancel', async () => {
        // An action without run closes the panel
        const panel = new PanelController({ actions: [{ label: 'Cancel' }] });
        panel.open('one');
        assert.equal(await panel.run(panel.current.actions[0]), true);
        assert.ok(!panel.isOpen);
    });

    await t.test('fail', async () => {
        // An action which fails leaves the panel open, and says why
        const save: PanelAction = { label: 'Save', run: async () => { throw new Error('invalid'); } };
        const panel = new PanelController({ actions: [save] });
        const errors: string[] = [];
        panel.onError((error, action) => errors.push(`${action.label}: ${(error as Error).message}`));
        panel.open('one');
        assert.equal(await panel.run(save), false);
        assert.ok(panel.isOpen);
        assert.ok(!panel.running);
        assert.deepEqual(errors, ['Save: invalid']);
    });

    await t.test('stay', async () => {
        // An action which doesn't close the panel, and one which is
        // disabled, or isn't the panel's, which doesn't run
        let runs = 0;
        const refresh: PanelAction = { label: 'Refresh', kind: 'ghost', close: false, run: () => { runs++; } };
        const disabled: PanelAction = { label: 'Delete', kind: 'danger', disabled: true, run: () => { runs++; } };
        const panel = new PanelController({ actions: [refresh, disabled] });
        panel.open('one');
        assert.equal(await panel.run(refresh), true);
        assert.equal(await panel.run(disabled), false);
        assert.equal(await panel.run({ label: 'Other', run: () => { runs++; } }), false);
        assert.equal(runs, 1);
        assert.ok(panel.isOpen);

        // The actions can be changed
        panel.setActions([{ label: 'Done' }]);
        assert.deepEqual(panel.current.actions.map((action) => action.label), ['Done']);
    });

    await t.test('replaced while running', async () => {
        // An action which replaces the panel's actions while it runs, such as
        // Add, which asks to sign in first, of Sign in, isn't running once
        // it's done, which the listeners are told, so the action which
        // replaced it can be run
        const signIn: PanelAction = { label: 'Sign in', close: false, run: () => {} };
        const panel = new PanelController<string>();
        const add: PanelAction = { label: 'Add', close: false, run: async () => { panel.setActions([signIn]); } };
        panel.setActions([add]);
        panel.open('one');
        const running: boolean[] = [];
        panel.subscribe(() => running.push(panel.running));
        assert.equal(await panel.run(add), true);
        assert.equal(panel.running, false);
        assert.equal(running.at(-1), false, 'the last state the listeners are told of isn\'t running');
        assert.deepEqual(panel.current.actions.map((action) => [action.label, action.running]), [['Sign in', false]]);
        assert.equal(await panel.run(signIn), true);
    });

    await t.test('tabs', async () => {
        // A panel of tabs has the actions of the tab which is chosen, which is
        // the first which isn't disabled, by default
        const mount: PanelAction = { label: 'Mount', run: () => undefined };
        const save: PanelAction = { label: 'Save', run: () => undefined };
        const remove: PanelAction = { label: 'Delete', kind: 'danger', disabled: true, run: () => undefined };
        const tabs: PanelTab[] = [
            { value: 'none', label: 'None', disabled: true },
            { value: 'view', label: 'View', actions: [mount] },
            { value: 'edit', label: 'Edit', actions: [save] },
            { value: 'delete', label: 'Delete', actions: [remove] },
        ];
        const panel = new PanelController<string>({ tabs });
        assert.equal(panel.tab, 'view');
        assert.deepEqual(panel.current.tabs.map((tab) => tab.value), ['none', 'view', 'edit', 'delete']);
        assert.deepEqual(panel.current.actions.map((action) => action.label), ['Mount']);

        // Another tab is chosen, of its actions, but not one which is
        // disabled, or isn't a tab
        assert.ok(panel.selectTab('edit'));
        assert.equal(panel.tab, 'edit');
        assert.deepEqual(panel.current.actions.map((action) => action.label), ['Save']);
        assert.ok(!panel.selectTab('none'));
        assert.ok(!panel.selectTab('missing'));
        assert.equal(panel.tab, 'edit');

        // The actions which are set are the chosen tab's, so they're its when
        // it's chosen again, such as Delete, once it's enabled
        assert.ok(panel.selectTab('delete'));
        assert.ok(panel.current.actions[0].disabled);
        panel.setActions([{ ...remove, disabled: false }]);
        assert.ok(panel.selectTab('view'));
        assert.ok(panel.selectTab('delete'));
        assert.ok(!panel.current.actions[0].disabled);

        // Setting the tabs again keeps the chosen one, when it's one of them,
        // or chooses the one of the value
        panel.setTabs(tabs);
        assert.equal(panel.tab, 'delete');
        panel.setTabs(tabs, 'edit');
        assert.equal(panel.tab, 'edit');
        panel.setTabs([]);
        assert.equal(panel.tab, undefined);
        assert.deepEqual(panel.current.tabs, []);
    });

    await t.test('tabs while running', async () => {
        // Another tab can't be chosen while an action runs
        let resolve!: () => void;
        const save: PanelAction = { label: 'Save', close: false, run: () => new Promise<void>((done) => { resolve = done; }) };
        const panel = new PanelController<string>({ tabs: [{ value: 'edit', label: 'Edit', actions: [save] }, { value: 'view', label: 'View' }] });
        panel.open('one');
        const saving = panel.run(save);
        assert.ok(!panel.selectTab('view'));
        assert.equal(panel.tab, 'edit');

        // An action which is set while it runs keeps running
        panel.setActions([save]);
        assert.ok(panel.current.actions[0].running);
        resolve();
        assert.equal(await saving, true);
        assert.ok(!panel.running);
        assert.ok(panel.selectTab('view'));
        assert.deepEqual(panel.current.actions, []);
    });

    await t.test('tabs set while running', async () => {
        // An action which runs is running while the tabs are set again, of the
        // same, or others, which don't have it, so it can't run again, nor
        // can another, until it settles
        let resolve!: () => void;
        const save: PanelAction = { label: 'Save', close: false, run: () => new Promise<void>((done) => { resolve = done; }) };
        const other: PanelAction = { label: 'Other', close: false, run: () => undefined };
        const tabs: PanelTab[] = [{ value: 'edit', label: 'Edit', actions: [save] }];
        const panel = new PanelController<string>({ tabs });
        panel.open('one');
        const saving = panel.run(save);
        panel.setTabs(tabs);
        assert.ok(panel.running);
        assert.ok(panel.current.actions[0].running);
        assert.equal(await panel.run(save), false);
        panel.setTabs([{ value: 'other', label: 'Other', actions: [other] }]);
        assert.ok(panel.running);
        assert.ok(!panel.current.actions[0].running);
        assert.equal(await panel.run(other), false);
        assert.equal(panel.selectTab('other'), false);
        resolve();
        assert.equal(await saving, true);
        assert.ok(!panel.running);
        assert.equal(await panel.run(other), true);
    });
});
