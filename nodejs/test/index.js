'use strict';

const _ = require('lodash');
const child_process = require('child_process');
const FaissdbReplicaSet = require("index").ReplicaSet;
const N = 300;

const FAISSDB = '../server/faissdb';
const FAISSDB_CONFPATH = '../config/test';

const MAIN_RESULT =  [
[ 'k87', 1.0001120567321777 ],
  [ 'k91', 0.9999834299087524 ],
  [ 'k90', 0.9999766945838928 ],
  [ 'k89', 0.9999513626098633 ],
  [ 'k92', 0.9999333620071411 ],
  [ 'k88', 0.9999009966850281 ],
  [ 'k93', 0.9998493194580078 ],
  [ 'k86', 0.999740719795227 ],
  [ 'k94', 0.9997310042381287 ],
  [ 'k85', 0.999709963798523 ],
  [ 'k81', 0.9996604919433594 ],
  [ 'k95', 0.9995778203010559 ],
  [ 'k96', 0.9993893504142761 ],
  [ 'k82', 0.9992419481277466 ],
  [ 'k83', 0.9992172122001648 ],
  [ 'k97', 0.9991652369499207 ],
  [ 'k84', 0.9990770816802979 ],
  [ 'k98', 0.9989048838615417 ],
  [ 'k99', 0.9986081719398499 ],
  [ 'k100', 0.9982744455337524 ],
  [ 'k79', 0.9981250762939453 ],
  [ 'k103', 0.9981074333190918 ],
  [ 'k101', 0.9979034066200256 ],
  [ 'k78', 0.9977555274963379 ],
  [ 'k80', 0.9975571632385254 ],
  [ 'k102', 0.9974945187568665 ],
  [ 'k77', 0.9974765777587891 ],
  [ 'k75', 0.9972800612449646 ],
  [ 'k76', 0.9969075322151184 ],
  [ 'k105', 0.9963751435279846 ]
];

const I3_RESULT = [
  [ 'k87', 1.0001120567321777 ],
  [ 'k93', 0.9998493194580078 ],
  [ 'k81', 0.9996604919433594 ],
  [ 'k96', 0.9993893504142761 ],
  [ 'k84', 0.9990770816802979 ],
  [ 'k99', 0.9986081719398499 ],
  [ 'k78', 0.9977555274963379 ],
  [ 'k102', 0.9974945187568665 ],
  [ 'k72', 0.9951022267341614 ],
  [ 'k108', 0.9942277669906616 ],
  [ 'k69', 0.9934375882148743 ],
  [ 'k111', 0.9920551776885986 ],
  [ 'k66', 0.9915408492088318 ],
  [ 'k63', 0.9893209934234619 ],
  [ 'k114', 0.9892469644546509 ],
  [ 'k117', 0.9865893125534058 ],
  [ 'k57', 0.9848154187202454 ],
  [ 'k54', 0.9822744727134705 ],
  [ 'k123', 0.979584276676178 ],
  [ 'k51', 0.9794931411743164 ],
  [ 'k48', 0.9766508936882019 ],
  [ 'k126', 0.9754918813705444 ],
  [ 'k129', 0.97098708152771 ],
  [ 'k42', 0.9704961180686951 ],
  [ 'k39', 0.967267632484436 ],
  [ 'k132', 0.9664472937583923 ],
  [ 'k36', 0.9639403820037842 ],
  [ 'k33', 0.9605231881141663 ],
  [ 'k138', 0.9541409611701965 ],
  [ 'k27', 0.9532678723335266 ]
];

const I15_RESULT = [
  [ 'k90', 0.9999766945838928 ],
  [ 'k75', 0.9972800612449646 ],
  [ 'k105', 0.9963751435279846 ],
  [ 'k60', 0.9872412085533142 ],
  [ 'k120', 0.9832820892333984 ],
  [ 'k45', 0.9738624095916748 ],
  [ 'k135', 0.9608237147331238 ],
  [ 'k30', 0.9570244550704956 ],
  [ 'k15', 0.9385786056518555 ],
  [ 'k150', 0.928156852722168 ],
  [ 'k0', 0.9191365242004395 ],
  [ 'k165', 0.8889648914337158 ],
  [ 'k180', 0.83761066198349 ],
  [ 'k195', 0.7823975682258606 ],
  [ 'k210', 0.7241705656051636 ],
  [ 'k225', 0.6643639206886292 ],
  [ 'k240', 0.604846179485321 ],
  [ 'k255', 0.5476592183113098 ],
  [ 'k270', 0.4930126368999481 ],
  [ 'k285', 0.4416840970516205 ]
];

const I9_RESULT = [
  [ 'k90', 0.9999766945838928 ],
  [ 'k81', 0.9996604919433594 ],
  [ 'k99', 0.9986081719398499 ],
  [ 'k72', 0.9951022267341614 ],
  [ 'k108', 0.9942277669906616 ],
  [ 'k63', 0.9893209934234619 ],
  [ 'k117', 0.9865893125534058 ],
  [ 'k54', 0.9822744727134705 ],
  [ 'k126', 0.9754918813705444 ],
  [ 'k45', 0.9738624095916748 ]
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
          reject(`Error code=${code} ${stdErrData}`);
        }
      });
    } catch(e) {
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
            info: console.log
          }
        });
        this.faissdbClient.init();
        this.delKeys = [];
        this.inputs = [];
        for(let i = 0; i < N; i++){
          let key = getKey(i);
          let collections = [];
          if(i%3 == 0) {
            collections = ['i3'];
          }
          if(i%15 == 0) {
            collections = ['i15'];
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
              collections: ['i9'],
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
              description: 'IVF2,PQ2_8',
              metric: 'InnerProduct',
              nprobe: 10,
              dimension: 2,
              syncinterval: 60000
            },
            dbs: [{
              collection: 'main', ntotal: 0
            }]});
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
          let [keys, distances] = await this.faissdbClient.primary.search('', 10, normalize([30, 70]));
          expect(keys).to.deep.equals([]);
          await this.faissdbClient.train(1);
          [keys, distances] = await this.faissdbClient.primary.search('', 10, normalize([30, 70]));
          expect(_.zip(keys, distances)).to.deep.equals(MAIN_RESULT.slice(0, 10));
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
          let [keys, distances] = await this.faissdbClient.search('', 10, normalize([30, 70]));
          expect(_.zip(keys, distances)).to.deep.equals(MAIN_RESULT.slice(0, 10));
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
          expect(_.zip(keys, distances)).to.deep.equals(I3_RESULT.slice(0, 10));
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
          expect(_.zip(keys, distances)).to.deep.equals(I15_RESULT.slice(0, 10));
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
            let [keys, distances] = await this.faissdbClient.search('', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(MAIN_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.search('i3', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(I3_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.search('i15', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(I15_RESULT, r => isInvalid(r[0])).slice(0,10));
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
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(MAIN_RESULT, r => isMainInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i3', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(I3_RESULT, r => isInvalid(r[0])).slice(0,10));
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i9', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(I9_RESULT);
          }
          {
            let [keys, distances] = await this.faissdbClient.secondaries[1].search('i15', 10, normalize([30, 70]));
            expect(_.zip(keys, distances)).to.deep.equals(_.reject(I15_RESULT, r => isInvalid(r[0])).slice(0,10));
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
