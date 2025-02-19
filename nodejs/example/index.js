const ReplicaSet = require('index').ReplicaSet;
const _ = require('lodash');

const normalize = (vector) => {
  let l = Math.pow(_.reduce(vector, (r, v) => r + v*v, 0), 0.5);
  return vector.map(v => v/l);
};

(async () => {
  let client = new ReplicaSet({
    connects: [{
      host: 'localhost',
      port: 20021,
    }]
  });
  client.init();
  let datas = [];
  for(let i = 0; i <= 1000; i++) {
    let v = normalize([i, 1000-i]);
    datas.push({
      key: `d${i}`,
      v: v,
      collection: ['main'],
    });
    datas.push({
      key: `s${i}`,
      v: {
        '0': v[0],
        '1': v[1],
      },
      collections: ['main'],
    });
  }
  console.log('dropall');
  await client.dropall();
  console.log('set');
  await client.set(datas);
  console.log('train');
  await client.train(0.7, true);
  console.log('del');
  await client.del(['k1', 'k10', 'k20']);
  console.log('search');
  {
    let [keys, distances] = await client.search('main', 20, normalize([0.1, 0.9]));
    console.log('Expects around k100', _.zip(keys, distances));
  }
  {
    let v = normalize([1, 1]);
    let [keys, distances] = await client.search('main', 20, {
      '0': v[0],
      '1': v[1],
    });
    console.log('Expects around k500', _.zip(keys, distances));
  }
  process.exit(0);
})();
