import { test } from 'node:test';
import assert from 'node:assert/strict';
import { parseRoute } from '../src/router/index.ts';

test('parseRoute', () => {
    // A route is the hash's page, and its parameters, which are decoded,
    // or the default page, when it's empty
    assert.deepEqual(parseRoute('', 'home'), { page: 'home', params: [] });
    assert.deepEqual(parseRoute('#', 'home'), { page: 'home', params: [] });
    assert.deepEqual(parseRoute('#roles', 'home'), { page: 'roles', params: [] });
    assert.deepEqual(parseRoute('#database/my%20db/public', 'home'), { page: 'database', params: ['my db', 'public'] });

    // A query is ignored, and an element which can't be decoded is kept
    assert.deepEqual(parseRoute('#roles?x=1', 'home'), { page: 'roles', params: [] });
    assert.deepEqual(parseRoute('#role/%E0', 'home'), { page: 'role', params: ['%E0'] });
});
