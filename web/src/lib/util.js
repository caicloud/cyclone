import moment from 'moment';
import qs from 'query-string';
import {
  STAGE,
  SPECIAL_EDGE_TYPE,
} from '@/components/workflow/component/add/graph-config';
export function FormatTime(time, formatStr = 'YYYY-MM-DD HH:mm:ss') {
  return moment(time).format(formatStr);
}

export const renderTemplate = data => {
  let domArr = [];
  const render = (object, parent) => {
    _.forEach(object, (v, k) => {
      const name = parent ? `${parent}.${k}` : k;
      if (_.isObject(v) && !_.isArray(v)) {
        render(v, name);
      } else {
        const dom = 'test';
        if (dom) {
          domArr.push(dom);
        }
      }
    });
  };
  render(data);
  return domArr;
};

export const getQuery = str => {
  const query = qs.parse(str);
  return query;
};

export const tranformStage = (stages, position) => {
  let nodes = [];
  let edges = [];
  const _position = _.values(JSON.parse(position));
  _.forEach(stages, (v, k) => {
    const pos = _position.find(p => p.title === v.name);
    const node = {
      id: `stage_${k}`,
      title: v.name,
      type: STAGE,
      ..._.pick(pos, ['x', 'y']),
    };
    nodes.push(node);
    if (_.isArray(v.depends)) {
      const edge = _.map(v.depends, d => {
        const index = _.findIndex(stages, s => {
          return s.name === d;
        });
        return {
          source: `stage_${index}`,
          target: `stage_${k}`,
          type: SPECIAL_EDGE_TYPE,
        };
      });
      edges = _.concat(edges, edge);
    }
  });
  return { nodes, edges };
};
