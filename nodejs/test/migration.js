'use strict';

const _ = require('lodash');
const child_process = require('child_process');
const FaissdbReplicaSet = require("index").ReplicaSet;
const FaissdbClient = require("index").Client;
const N = 300;

const sleep = (ms) => new Promise((resolve) => setTimeout(() => resolve(), ms));

const normalize = (vector) => {
  let l = Math.pow(_.reduce(vector, (r, v) => r + v*v, 0), 0.5);
  return vector.map(v => v/l);
};

const getKey = (i) => {
  return `k${i}`;
};

// Containers are pre-created by ci/test_migration.sh; the suite switches node
// versions by starting/stopping them through the docker API. Nodes are
// stopped with SIGTERM directly (graceful shutdown flushing the FAISS
// indexes) because images up to 0.4.0 declared STOPSIGNAL SIGRTMIN+3, which
// the server does not handle, making a plain `docker stop` an abrupt kill.
const dockerCtl = (action, name) => {
  child_process.execSync(`curl -s --unix-socket /var/run/docker.sock -X POST "http://localhost/containers/${name}/${action}"`);
};
const nodeRunning = (name) => {
  let out = child_process.execSync(`curl -s --unix-socket /var/run/docker.sock "http://localhost/containers/${name}/json"`).toString();
  return JSON.parse(out).State.Running;
};
const startNode = (name) => dockerCtl('start', name);
const stopNode = async (name) => {
  dockerCtl('kill?signal=SIGTERM', name);
  while(nodeRunning(name)) {
    await sleep(500);
  }
};

// faissdb >= 0.5.0 keeps index files under <dbpath>/indexes/; the documented
// migration is a manual mv while the node is stopped. ci/test_migration.sh
// mounts the data volumes on this driver at /mig<n>-data.
const upgradeIndexLayout = (n) => {
  child_process.execSync(`mkdir -p /mig${n}-data/indexes && find /mig${n}-data -maxdepth 1 -type f -exec mv {} /mig${n}-data/indexes/ \\;`);
};
const rollbackIndexLayout = (n) => {
  child_process.execSync(`find /mig${n}-data/indexes -maxdepth 1 -type f -exec mv {} /mig${n}-data/ \\;`);
};

const NODE1_OLD = 'faissdb-mig-node1-old';
const NODE1_NEW = 'faissdb-mig-node1-new';
const NODE2_OLD = 'faissdb-mig-node2-old';
const NODE2_NEW = 'faissdb-mig-node2-new';

const waitReady = async (client) => {
  while(true) {
    let status = await client.status();
    if(status.status == 100) {
      return;
    }
    await sleep(500);
  }
};

const waitForDbstats = async (client, predicate) => {
  while(true) {
    try {
      let dbStats = await client.dbstats();
      if(predicate(dbStats)) {
        return dbStats;
      }
    } catch(e) {
      console.log(`waitForDbstats() retry: ${e}`);
    }
    await sleep(500);
  }
};

const countsEqual = (dbStats, expected) => {
  let actual = {};
  for(let db of dbStats.dbs) {
    actual[db.collection] = db.ntotal;
  }
  return _.isEqual(actual, expected);
};

// 0.3.x cannot flush its FAISS indexes on shutdown (its graceful-shutdown
// path exits before the flush; fixed in 0.4.1), and a freshly full-synced
// secondary has nothing in its local oplog to replay the base data from.
// Before stopping a 0.3.x node, wait until the periodic sync has persisted
// the indexes past the current oplog position — the same step a real
// 0.3.x -> 0.4.0 upgrade must perform.
const waitFlushed = async (client) => {
  let target = (await client.dbstats()).lastkey;
  await waitForDbstats(client, (dbStats) => dbStats.lastsynced != '' && dbStats.lastsynced >= target);
};

describe('migration', ()=> {
  describe('0.3.x <-> 0.4.0', ()=> {
    before(() => {
      this.rs = new FaissdbReplicaSet({
        connects: [{
          host: "localhost",
          port: 20021
        }, {
          host: "localhost",
          port: 20022
        }],
        prepareInterval: 1000,
        logger: {
          info: console.log,
          error: console.log,
        }
      });
      this.rs.init();
      this.node1 = new FaissdbClient({connect: {host: "localhost", port: 20021}});
      this.node1.init();
      this.node2 = new FaissdbClient({connect: {host: "localhost", port: 20022}});
      this.node2.init();
      this.searchEquals = async () => {
        let [keys1] = await this.node1.search('main', 10, normalize([30, 70]));
        let [keys2] = await this.node2.search('main', 10, normalize([30, 70]));
        expect(keys1.length).to.equals(10);
        expect(keys1).to.deep.equals(keys2);
      };
      this.expectedCounts = {};
      this.waitCluster = async (expected) => {
        await waitForDbstats(this.node1, (dbStats) => countsEqual(dbStats, expected));
        await waitForDbstats(this.node2, (dbStats) => countsEqual(dbStats, expected));
      };
      this.waitPrimary = async () => {
        while(true) {
          await this.rs._prepare();
          if(this.rs.primary && this.rs.primary.isReady()) {
            return;
          }
          await sleep(1000);
        }
      };
    });

    it('Build 0.3.1 cluster', () => {
      return new Promise(async (resolve, reject) => {
        try {
          startNode(NODE1_OLD);
          startNode(NODE2_OLD);
          await sleep(3000);
          child_process.execSync(`curl -s http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "localhost:21021", "primary": true}, {"id": 2, "host": "localhost:21022", "primary": false}]}'`);
          await waitReady(this.node1);
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Load data on 0.3.1', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let inputs = [];
          for(let i = 0; i < N; i++) {
            let collections = ['main'];
            if(i%2 == 1) {
              collections.push('odd');
            }
            inputs.push({key: getKey(i), v: normalize([i, N-i]), collections: collections});
          }
          let [nStored, nErrors] = await this.rs.set(inputs);
          expect(nStored).to.equals(N);
          expect(nErrors).to.equals(0);
          await this.rs.train(1);
          this.expectedCounts = {main: 300, odd: 150};
          await this.waitCluster(this.expectedCounts);
          await this.searchEquals();
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Upgrade secondary to 0.4.0', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await waitFlushed(this.node2);
          await stopNode(NODE2_OLD);
          upgradeIndexLayout(2);
          startNode(NODE2_NEW);
          await waitReady(this.node2);
          await this.waitCluster(this.expectedCounts);

          let updates = [];
          for(let i = 0; i < N; i++) {
            if(i%5 == 0) {
              let collections = ['main'];
              if(i%2 == 1) {
                collections.push('odd');
              }
              updates.push({key: getKey(i), v: normalize([i + 0.5, N-i]), collections: collections});
            }
          }
          await this.rs.set(updates);
          let delKeys = [];
          for(let i = 0; i < N; i++) {
            if(i%15 == 0) {
              delKeys.push(getKey(i));
            }
          }
          await this.rs.del(delKeys);
          this.expectedCounts = {main: 280, odd: 140};
          await this.waitCluster(this.expectedCounts);
          await this.searchEquals();
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Upgrade primary to 0.4.0', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await waitFlushed(this.node1);
          await stopNode(NODE1_OLD);
          upgradeIndexLayout(1);
          startNode(NODE1_NEW);
          await waitReady(this.node1);
          await this.waitCluster(this.expectedCounts);
          await this.waitPrimary();

          let newInputs = [];
          for(let i = 0; i < 10; i++) {
            newInputs.push({key: `n${i}`, v: normalize([i + 0.25, N-i]), collections: ['main']});
          }
          await this.rs.set(newInputs);
          let collectionsUpdates = [];
          for(let i = 0; i < N; i++) {
            if(i%4 == 0 && i%15 != 0) {
              collectionsUpdates.push({key: getKey(i), collections: ['main', 'four']});
            }
          }
          let [nStored, nErrors] = await this.rs.setCollections(collectionsUpdates);
          expect(nStored).to.equals(70);
          expect(nErrors).to.equals(0);
          this.expectedCounts = {four: 70, main: 290, odd: 140};
          await this.waitCluster(this.expectedCounts);
          await this.searchEquals();
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Roll back secondary to 0.3.1', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await stopNode(NODE2_NEW);
          rollbackIndexLayout(2);
          startNode(NODE2_OLD);
          await waitReady(this.node2);
          await this.waitCluster(this.expectedCounts);

          let collectionsUpdates = [];
          for(let i = 0; i < 10; i++) {
            collectionsUpdates.push({key: `n${i}`, collections: ['main', 'newcol']});
          }
          let [nStored, nErrors] = await this.rs.setCollections(collectionsUpdates);
          expect(nStored).to.equals(10);
          expect(nErrors).to.equals(0);
          this.expectedCounts = {four: 70, main: 290, newcol: 10, odd: 140};
          await this.waitCluster(this.expectedCounts);
          await this.searchEquals();
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Roll back primary to 0.3.1', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await stopNode(NODE1_NEW);
          rollbackIndexLayout(1);
          startNode(NODE1_OLD);
          await waitReady(this.node1);
          await this.waitCluster(this.expectedCounts);
          await this.waitPrimary();

          let err = null;
          try {
            await this.rs.setCollections([{key: 'n0', collections: ['main']}]);
          } catch(e) {
            err = e;
          }
          expect(String(err)).to.match(/UNIMPLEMENTED|not implemented/i);

          let newInputs = [];
          for(let i = 0; i < 10; i++) {
            newInputs.push({key: `p${i}`, v: normalize([i + 0.75, N-i]), collections: ['main']});
          }
          await this.rs.set(newInputs);
          this.expectedCounts = {four: 70, main: 300, newcol: 10, odd: 140};
          await this.waitCluster(this.expectedCounts);
          await this.searchEquals();
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

  });
});
