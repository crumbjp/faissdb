'use strict';

const _ = require('lodash');

let Client = require('./client');
exports.Client = Client;

class ReplicaSet {
  constructor(options) {
    this.clients = _.map(options.connects, connect => new Client({connect: connect}));
    this.primary = null;
    this.secondaries = [];
    this.lastPrepare = null;
    this.prepareInterval = options.prepareInterval || 5000;
    this.debug = options.debug;
    this.logger = options.logger;
  }

  init() {
    _.each(this.clients, client => client.init());
  }

  log(msg) {
    if(this.logger) {
      this.logger.info(msg);
    }
  }

  _prepare() {
    return new Promise(async (resolve, reject) => {
      try {
        if(!this.lastPrepare || (this.lastPrepare + this.prepareInterval) < new Date().getTime()) {
          if(this.debug) {
            this.log('_prepare() check clients');
          }
          let promises = [];
          for(let client of this.clients) {
            promises.push(client.status());
          }
          try {
            await Promise.all(promises);
          } catch(e) {
            // Nothing to do
          }
          this.primary = null;
          this.secondaries = [];
          for(let client of this.clients) {
            if(client.isPrimary()) {
              this.primary = client;
            } else if(client.isSecondary()) {
              this.secondaries.push(client);
            }
          }
          if(this.debug) {
            this.log(`_prepare() clients ${_.map(this.clients, (client) => JSON.stringify(client.currentStatus))}`);
          }
          this.lastPrepare = new Date().getTime();
        }
        resolve();
      } catch(e) {
        reject(e);
      }
    });
  }

  set(inputs, options = {}) {
    return new Promise(async (resolve, reject) => {
      try {
        await this._prepare();
        if(!this.primary) {
          return reject('No primary found');
        }
        resolve(await this.primary.set(inputs, options));
      } catch(e) {
        reject(e);
      }
    });
  }

  train(proportion, options = {}) {
    return new Promise(async (resolve, reject) => {
      try {
        await this._prepare();
        if(!this.primary) {
          return reject('No primary found');
        }
        resolve(await this.primary.train(proportion, options));
      } catch(e) {
        reject(e);
      }
    });
  }

  del(keys, options = {}) {
    return new Promise(async (resolve, reject) => {
      try {
        await this._prepare();
        if(!this.primary) {
          return reject('No primary found');
        }
        resolve(await this.primary.del(keys, options));
      } catch(e) {
        reject(e);
      }
    });
  }

  search(collection, n, v, options = {}) {
    return new Promise(async (resolve, reject) => {
      try {
        await this._prepare();
        let client = _.chain(this.secondaries).filter(secondary => secondary.isReady()).sample().value();
        client = client || (this.primary.isReady() ? this.primary : null);
        if(!client) {
          return reject('No valid node found');
        }
        if(this.debug) {
          this.log(`search() from ${JSON.stringify(client.currentStatus)}`);
        }
        resolve(await client.search(collection, n, v, options));
      } catch(e) {
        reject(e);
      }
    });
  }

  dropall() {
    return new Promise(async (resolve, reject) => {
      try {
        await this._prepare();
        if(!this.primary) {
          return reject('No primary found');
        }
        resolve(await this.primary.dropall());
      } catch(e) {
        reject(e);
      }
    });
  }
}

exports.ReplicaSet = ReplicaSet;
