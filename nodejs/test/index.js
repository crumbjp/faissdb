'use strict';

const _ = require('lodash');
const child_process = require('child_process');
const FaissdbReplicaSet = require("index").ReplicaSet;
const N = 300;

const FAISSDB = '../server/faissdb';
const FAISSDB_CONFPATH = '../config/test';

const MAIN_RESULT =  [
  [ 'k87',  1.00011],
  [ 'k91',  0.99998],
  [ 'k90',  0.99997],
  [ 'k89',  0.99995],
  [ 'k92',  0.99993],
  [ 'k88',  0.99990],
  [ 'k93',  0.99984],
  [ 'k86',  0.99974],
  [ 'k94',  0.99973],
  [ 'k85',  0.99970],
  [ 'k81',  0.99966],
  [ 'k95',  0.99957],
  [ 'k96',  0.99938],
  [ 'k82',  0.99924],
  [ 'k83',  0.99921],
  [ 'k97',  0.99916],
  [ 'k84',  0.99907],
  [ 'k98',  0.99890],
  [ 'k99',  0.99860],
  [ 'k100', 0.99827],
  [ 'k79',  0.99812],
  [ 'k103', 0.99810],
  [ 'k101', 0.99790],
  [ 'k78',  0.99775],
  [ 'k80',  0.99755],
  [ 'k102', 0.99749],
  [ 'k77',  0.99747],
  [ 'k75',  0.99728],
  [ 'k76',  0.99690],
  [ 'k105', 0.99637],
];

const I3_RESULT = [
  [ 'k87',  1.00011],
  [ 'k93',  0.99984],
  [ 'k81',  0.99966],
  [ 'k96',  0.99938],
  [ 'k84',  0.99907],
  [ 'k99',  0.99860],
  [ 'k78',  0.99775],
  [ 'k102', 0.99749],
  [ 'k72',  0.99510],
  [ 'k108', 0.99422],
  [ 'k69',  0.99343],
  [ 'k111', 0.99205],
  [ 'k66',  0.99154],
  [ 'k63',  0.98932],
  [ 'k114', 0.98924],
  [ 'k117', 0.98658],
  [ 'k57',  0.98481],
  [ 'k54',  0.98227],
  [ 'k123', 0.97958],
  [ 'k51',  0.97949],
  [ 'k48',  0.97665],
  [ 'k126', 0.97549],
  [ 'k129', 0.97098],
  [ 'k42',  0.97049],
  [ 'k39',  0.96726],
  [ 'k132', 0.96644],
  [ 'k36',  0.96394],
  [ 'k33',  0.96052],
  [ 'k138', 0.95414],
  [ 'k27',  0.95326],
];

const I15_RESULT = [
  [ 'k90',  0.99997],
  [ 'k75',  0.99728],
  [ 'k105', 0.99637],
  [ 'k60',  0.98724],
  [ 'k120', 0.98328],
  [ 'k45',  0.97386],
  [ 'k135', 0.96082],
  [ 'k30',  0.95702],
  [ 'k15',  0.93857],
  [ 'k150', 0.92815],
  [ 'k0',   0.91913],
  [ 'k165', 0.88896],
  [ 'k180', 0.83761],
  [ 'k195', 0.78239],
  [ 'k210', 0.72417],
  [ 'k225', 0.66436],
  [ 'k240', 0.60484],
  [ 'k255', 0.54765],
  [ 'k270', 0.49301],
  [ 'k285', 0.44168],
];

const I9_RESULT = [
  [ 'k90',  0.99997],
  [ 'k81',  0.99966],
  [ 'k99',  0.99860],
  [ 'k72',  0.99510],
  [ 'k108', 0.99422],
  [ 'k63',  0.98932],
  [ 'k117', 0.98658],
  [ 'k54',  0.98227],
  [ 'k126', 0.97549],
  [ 'k45',  0.97386],
];

const normalize = (vector) => {
  let l = Math.pow(_.reduce(vector, (r, v) => r + v*v, 0), 0.5);
  return vector.map(v => v/l);
};

const sleep = (ms) => new Promise((resolve) => setTimeout(() => resolve(), ms));

const getKey = (i) => {
  return `k${i}`;
};

const cmd = async (cmd, cwd = process.cwd(), extraEnv = {}, mode = {}) => {
  return new Promise(async (resolve, reject) => {
    try {
      var options = {
        cwd: cwd,
        env: _.merge({}, process.env, extraEnv),
        shell: '/bin/bash'
      };
      console.log(`[cmd] ${cmd}`);
      let forked = child_process.spawn(cmd, options);
      forked.stdin.end();
      let stdOutData = '';
      let stdErrData = '';
      if(!mode.noPipe) {
        forked.stdout.on('data', (data) => {
          if(mode.simple) {
            stdOutData = data;
            process.stdout.write(data);
          } else {
            stdOutData += String(data);
          }
        });
        forked.stderr.on('data', (data) => {
          if(mode.simple) {
            stdErrData = data;
            process.stderr.write(data);
          } else {
            stdErrData += String(data);
          }
        });
      }
      forked.on('close', (code) => {
        if(code == 0) {
          console.log(`[result] ${stdOutData}`);
          resolve(stdOutData);
        } else {
          console.log(`Error code=${code} ${stdErrData}`);
          reject(`Error code=${code} ${stdErrData}`);
        }
      });
    } catch(e) {
      console.log(e);
      reject(e);
    }
  });
};

describe('index', ()=> {
  describe('faissdb', ()=> {
    before(() => {
      return new Promise(async (resolve, reject) => {
        this.faissdbClient = new FaissdbReplicaSet({
          connects: [{
            host: "localhost",
            port: 20021
          }, {
            host: "localhost",
            port: 20022
          }, {
            host: "localhost",
            port: 20023
          }],
          debug: true,
          logger: {
            info: console.log,
            error: console.log,
          }
        });
        this.faissdbClient.init();
        this.delKeys = [];
        this.inputs = [];
        for(let i = 0; i < N; i++) {
          let key = getKey(i);
          let collections = ['main'];
          if(i%3 == 0) {
            collections = ['main', 'i3'];
          }
          if(i%15 == 0) {
            collections = ['main', 'i15'];
          }
          if(i%2 == 0) {
            this.delKeys.push(key);
          }
          this.inputs.push({
            key: key,
            v: normalize([i, N-i]),
            collections: collections,
          });
        }
        this.updates = [];
        for(let i = 0; i < N; i++){
          if(i%9 == 0) {
            this.updates.push({
              key: getKey(i),
              v: normalize([i, N-i]),
              collections: ['main', 'i9'],
            });
          }
        }
        resolve(true);
      });
    });

    after(() => {
      return new Promise(async (resolve, reject) => {
        resolve(true);
      });
    });

    it('Build cluster', () => {
      return new Promise(async (resolve, reject) => {
        try {
          cmd(`${FAISSDB} ${FAISSDB_CONFPATH}/config1.yml`);
          cmd(`${FAISSDB} ${FAISSDB_CONFPATH}/config2.yml`);
          await sleep(3000);
          await cmd(`curl -v http://localhost:9091/replicaset -XPUT -d '{"replica": "rs", "members": [{"id": 1, "host": "localhost:21021", "primary": true}, {"id": 2, "host": "localhost:21022", "primary": false}, {"id": 3, "host": "localhost:21023", "primary": false}]}'`);
          await sleep(3000);
          await this.faissdbClient._prepare();
          let primaryStatus = await this.faissdbClient.primary.status();
          expect(primaryStatus).to.deep.equals({
             id: 1, status: 100, role: 1
          });
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(primaryDbStats).to.deep.equals({
            istrained: false,
            lastsynced: '',
            lastkey: '',
            status: 100,
            faissConfig: {
              description: 'IVF2,PQ2x8',
              metric: 'InnerProduct',
              nprobe: 10,
              dimension: 2,
              syncinterval: 60000
            },
            dbs: []});
          expect(this.faissdbClient.secondaries.length).to.equals(1);
          let secondaryStatus = await this.faissdbClient.secondaries[0].status();
          expect(secondaryStatus).to.deep.equals({
             id: 2, status: 10, role: 2
          });
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Put first data', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let [nStored, nErrors] = await this.faissdbClient.set(this.inputs);
          expect(nStored).to.equals(N);
          expect(nErrors).to.equals(0);
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(_.sortBy(primaryDbStats.dbs, 'collection')).to.deep.equals([{
            collection: 'i15', ntotal: 0
          }, {
            collection: 'i3', ntotal: 0
          }, {
            collection: 'main', ntotal: 0
          }]);
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Train', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let [emptyKeys, emptyDistances] = await this.faissdbClient.primary.search('main', 10, normalize([30, 70]));
          expect(emptyKeys).to.deep.equals([]);
          await this.faissdbClient.train(1);
          let [keys, distances] = await this.faissdbClient.primary.search('main', 10, normalize([30, 70]));
          expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(MAIN_RESULT.slice(0, 10));
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(_.sortBy(primaryDbStats.dbs, 'collection')).to.deep.equals([{
            collection: 'i15', ntotal: 20
          }, {
            collection: 'i3', ntotal: 80
          }, {
            collection: 'main', ntotal: 300
          }]);
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('WaitFor secondary', () => {
      return new Promise(async (resolve, reject) => {
        try {
          while(true) {
            let secondaryStatus = await this.faissdbClient.secondaries[0].status();
            await sleep(500);
            if(secondaryStatus.status == 100) {
              break;
            }
          }
          let secondaryDbStats = await this.faissdbClient.secondaries[0].dbstats();
          expect(_.sortBy(secondaryDbStats.dbs, 'collection')).to.deep.equals([{
            collection: 'i15', ntotal: 20
          }, {
            collection: 'i3', ntotal: 80
          }, {
            collection: 'main', ntotal: 300
          }]);
          let [keys, distances] = await this.faissdbClient.search('main', 10, normalize([30, 70]));
          expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(MAIN_RESULT.slice(0, 10));
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Search i3', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let [keys, distances] = await this.faissdbClient.search('i3', 10, normalize([30, 70]));
          expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(I3_RESULT.slice(0, 10));
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Search i15', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let [keys, distances] = await this.faissdbClient.search('i15', 10, normalize([30, 70]));
          expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(I15_RESULT.slice(0, 10));
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Delete', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await this.faissdbClient.del(this.delKeys);
          let expectedDbs = [{
            collection: 'i15', ntotal: 10
          }, {
            collection: 'i3', ntotal: 40
          }, {
            collection: 'main', ntotal: 150
          }];
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(_.sortBy(primaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
          while(true) {
            let secondaryDbStats = await this.faissdbClient.secondaries[0].dbstats();
            if(_.find(secondaryDbStats.dbs, db => db.collection == 'main').ntotal == 150) {
              expect(_.sortBy(secondaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
              break;
            }
            await sleep(500);
          }
          let isInvalid = (key) => {
            return this.delKeys.indexOf(key) >=0;
          };
          {
            let [keys, distances] = await this.faissdbClient.search('main', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(MAIN_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.search('i3', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(I3_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.search('i15', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(I15_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Start new node', () => {
      return new Promise(async (resolve, reject) => {
        try {
          cmd(`${FAISSDB} ${FAISSDB_CONFPATH}/config3.yml`);
          while(true) {
            await this.faissdbClient._prepare();
            if(this.faissdbClient.secondaries.length == 2) {
              break;
            }
            await sleep(500);
          }
          while(true) {
            let secondaryDbStats = await this.faissdbClient.secondaries[1].dbstats();
            if(_.find(secondaryDbStats.dbs, db => db.collection == 'main').ntotal == 150) {
              break;
            }
          }
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Update', () => {
      return new Promise(async (resolve, reject) => {
        try {
          let [nStored, nErrors] = await this.faissdbClient.set(this.updates);
          expect(nStored).to.equals(34);
          expect(nErrors).to.equals(0);
          let expectedDbs = [{
            collection: 'i15', ntotal: 7
          }, {
            collection: 'i3', ntotal: 26
          }, {
            collection: 'i9', ntotal: 34
          }, {
            collection: 'main', ntotal: 167
          }];
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(_.sortBy(primaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
          while(true) {
            let secondaryDbStats = await this.faissdbClient.secondaries[1].dbstats();
            if(_.find(secondaryDbStats.dbs, db => db.collection == 'main').ntotal == 167) {
              expect(_.sortBy(secondaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
              break;
            }
            await sleep(500);
          }
          let isMainInvalid = (key) => {
            return this.delKeys.indexOf(key) >= 0 && _.map(this.updates, r => r.key).indexOf(key) < 0;
          };
          let isInvalid = (key) => {
            return this.delKeys.indexOf(key) >= 0 || _.map(this.updates, r => r.key).indexOf(key) >= 0;
          };
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('main', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(MAIN_RESULT, r => isMainInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i3', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(I3_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i9', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(I9_RESULT);
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i15', 10, normalize([30, 70]));
            expect(_.zip(keys, _.map(distances, (distance) => _.floor(distance, 5)))).to.deep.equals(_.reject(I15_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

    it('Dropall', () => {
      return new Promise(async (resolve, reject) => {
        try {
          await this.faissdbClient.dropall();
          let expectedDbs = [{
            collection: 'i15', ntotal: 0
          }, {
            collection: 'i3', ntotal: 0
          }, {
            collection: 'i9', ntotal: 0
          }, {
            collection: 'main', ntotal: 0
          }];
          let primaryDbStats = await this.faissdbClient.primary.dbstats();
          expect(_.sortBy(primaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
          while(true) {
            let secondaryDbStats = await this.faissdbClient.secondaries[0].dbstats();
            if(_.find(secondaryDbStats.dbs, db => db.collection == 'main').ntotal == 0) {
              expect(_.sortBy(secondaryDbStats.dbs, 'collection')).to.deep.equals(expectedDbs);
              break;
            }
            await sleep(500);
          }
          resolve();
        } catch(e) {
          reject(e);
        }
      });
    });

  });
});
