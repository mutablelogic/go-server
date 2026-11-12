// The broadcaster's schema, which mirrors go-server's broadcaster/schema

// ChangeAction is how something changed: it was created, updated or deleted
export type ChangeAction = 'create' | 'update' | 'delete';

// Change is how something changed, which is its action, what changed, which
// is a table of a schema, and which row, or rows, changed, when it says,
// such as by an identifier which the rows share, and what the table says of
// the change, such as a row's new serial, when it says anything
export interface Change {
    action: ChangeAction;
    schema?: string;
    table?: string;
    row?: string;
    data?: unknown;
}

// ChangeFilter matches changes: those of its schema, table and action, each
// of which matches any when it's undefined
export interface ChangeFilter {
    schema?: string;
    table?: string;
    action?: ChangeAction;
}

// The names of the server-sent events of a stream of changes: each change,
// and that the server is stopping, before the stream ends
export const ChangeName = 'change';
export const ChangeShutdown = 'shutdown';

// changeMatches returns true when the change matches the filter
export function changeMatches(filter: ChangeFilter, change: Change): boolean {
    return (!filter.schema || filter.schema === change.schema) &&
        (!filter.table || filter.table === change.table) &&
        (!filter.action || filter.action === change.action);
}
