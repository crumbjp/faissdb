'use strict';

const _ = require('lodash');

const FeatureGrpc = require('./grpc_feature/feature_grpc_pb');
const Feature = require('./grpc_feature/feature_pb');
const FeatureClient = FeatureGrpc.FeatureClient;
const GrpcJs = require('@grpc/grpc-js');

const ROLE_PRIMARY = 1;
const ROLE_SECONDARY = 2;

const STATUS_NONE = 0;
const STATUS_READY = 100;

class Client {
  constructor(options) {
    this.host = options.connect.host;
    this.port = options.connect.port;
    this.currentStatus = {};
  }

  init() {
    this.client = new FeatureClient(
      `${this.host}:${this.port}`,
      GrpcJs.credentials.createInsecure(),
      {
        "grpc.keepalive_time_ms": 3000,
        "grpc.keepalive_timeout_ms": 1000,
        "grpc.keepalive_permit_without_calls": 1,
        "grpc.max_send_message_length": 100*1024*1024,
      }
    );
  }

  _request(method, req) {
    return new Promise((resolve, reject) => {
      this.client[method](req, (err, resp) => {
        if(err) {
          if(err.code == 14) {
            this.currentStatus.status = STATUS_NONE;
          }
          return reject(err);
        }
        resolve(resp);
      });
    });
  }

  toSparse(obj) {
    return _.map(obj, (v, k) => `${k}:${obj[k]}`).join(',');
  }

  async status() {
    try {
      let statusRequest = new Feature.StatusRequest();
      let reply = await this._request('status', statusRequest);
      this.currentStatus = {
        id: reply.getId(),
        status: reply.getStatus(),
        role: reply.getRole()
      };
    } catch(e) {
      this.currentStatus.status = STATUS_NONE;
    }
    return this.currentStatus;
  }

  isPrimary() {
    return this.currentStatus.role == ROLE_PRIMARY;
  }

  isSecondary() {
    return this.currentStatus.role == ROLE_SECONDARY;
  }

  isReady() {
    return this.currentStatus.status == STATUS_READY;
  }

  /*
   * datas: [data]
   * data: {
   *   key: string
   *   v: `dense vector array` or `sparse vector object`
   *   collections: `[string] index names`
   * }
  */
  async set(inputs, options = {}) {
    let setRequest = new Feature.SetRequest();
    for(let input of inputs) {
      let data = new Feature.Data();
      data.setKey(input.key);
      if(_.isArray(input.v)) {
        data.setVList(input.v);
      } else {
        data.setSparsev(this.toSparse(input.v));
      }
      data.setCollectionsList(input.collections);
      setRequest.addData(data);
    }
    let reply = await this._request('set', setRequest);
    return [reply.getNstored(), reply.getNerror()];
  }

  async train(proportion, options = {}) {
    let trainRequest = new Feature.TrainRequest();
    trainRequest.setProportion(proportion);
    trainRequest.setForce(options.force);
    await this._request('train', trainRequest);
  }

  async del(keys, options = {}) {
    let delRequest = new Feature.DelRequest();
    delRequest.setKeyList(keys);
    await this._request('del', delRequest);
  }

  /*
   * v: `dense vector array` or `sparse vector object`
   * collections: string target index name
  */
  async search(collection, n, v, options = {}) {
    let searchRequest = new Feature.SearchRequest();
    searchRequest.setCollection(collection);
    searchRequest.setN(n);
    if(_.isArray(v)) {
      searchRequest.setVList(v);
    } else {
      searchRequest.setSparsev(this.toSparse(v));
    }
    let reply = await this._request('search', searchRequest);
    return [reply.getKeysList(), reply.getDistancesList()];
  }

  async dropall() {
    let dropallRequest = new Feature.DropallRequest();
    await this._request('dropall', dropallRequest);
  }

  async dbstats() {
    let dbStatsRequest = new Feature.DbStatsRequest();
    let reply = await this._request('dbStats', dbStatsRequest);
    let faissConfig = reply.getFaissconfig();
    return {
      istrained: reply.getIstrained(),
      lastsynced: reply.getLastsynced(),
      lastkey: reply.getLastkey(),
      status: reply.getStatus(),
      faissConfig: {
        description: faissConfig.getDescription(),
        metric: faissConfig.getMetric(),
        nprobe: faissConfig.getNprobe(),
        dimension: faissConfig.getDimension(),
        syncinterval: faissConfig.getSyncinterval(),
      },
      dbs: _.map(reply.getDbsList(), (db) => ({
        collection: db.getCollection(),
        ntotal: db.getNtotal(),
      }))
    };
  }
}

module.exports = Client;
