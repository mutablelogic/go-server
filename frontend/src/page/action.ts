// A page's actions, such as Create, which a page header shows as buttons on
// its right, and runs, when they're clicked

// PageAction is an action on the right of a page's header: its label, its
// kind, which is primary by default, its icon, a 16px Carbon icon
// descriptor, or a rendered icon, and what it runs, when it's clicked. When
// run returns a promise, its button is disabled until it settles.
export interface PageAction {
    label: string;
    kind?: 'primary' | 'secondary' | 'tertiary' | 'ghost' | 'danger';
    icon?: unknown;
    disabled?: boolean;
    run: () => void | Promise<unknown>;
}

// runAction runs an action, and calls busy with true while a promise it
// returns is pending, and then false, and onError with what it throws, or
// rejects with, rather than letting it escape. It returns once the action
// has settled.
export async function runAction(action: PageAction, busy: (busy: boolean) => void, onError: (error: unknown) => void): Promise<void> {
    let result: void | Promise<unknown>;
    try {
        result = action.run();
    } catch (error) {
        onError(error);
        return;
    }
    if (!(result instanceof Promise)) {
        return;
    }
    busy(true);
    try {
        await result;
    } catch (error) {
        onError(error);
    } finally {
        busy(false);
    }
}
