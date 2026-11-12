import { test } from 'node:test';
import assert from 'node:assert/strict';
import { PanelController, type PanelAction, type PanelState } from '../src/panel/index.ts';

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
});
