# faiss-db-client

Node.js client for [faissdb](https://github.com/crumbjp/faissdb), a replicated vector database built on FAISS and RocksDB.

## Install

```
npm install faiss-db-client
```

## Usage

```js
const { ReplicaSet } = require('faiss-db-client');

const client = new ReplicaSet({
  connects: [
    { host: 'localhost', port: 20021 },
    { host: 'localhost', port: 20022 },
    { host: 'localhost', port: 20023 },
  ],
});
client.init();

// Upsert vectors (writes go to the primary automatically)
const [nStored, nErrors, errorIndexes] = await client.set([
  { key: 'k1', v: [0.1, 0.2], collections: ['main', 'sub'] },
]);

// Update collection membership without resending the vector (faissdb >= 0.4.0)
await client.setCollections([
  { key: 'k1', collections: ['main'] },
]);

// Search a collection (reads are balanced to secondaries)
const [keys, distances] = await client.search('main', 10, [0.1, 0.2]);

// Delete (resolves to the indexes of keys that did not exist)
await client.del(['k1']);
```

Sparse vectors are also accepted: pass `v` as an object of `{ index: value }` instead of an array.

`set` and `setCollections` resolve to `[nStored, nErrors, errorIndexes]` where `errorIndexes` lists the 0-based positions of the inputs that failed (faissdb >= 0.4.3; older servers report an empty list).

Other operations: `train(proportion)`, `dropall()`, `dbstats()`, `status()`. A single-node `Client` class is exported as well.

## Reference

- [faissdb repository](https://github.com/crumbjp/faissdb)
- [API reference](https://github.com/crumbjp/faissdb/blob/main/REFERENCE.md)
- [Full example (E2E test)](https://github.com/crumbjp/faissdb/blob/main/nodejs/test/index.js)
