import moment from 'moment';
import qs from 'query-string';
import {
  STAGE,
  SPECIAL_EDGE_TYPE,
} from '@/components/workflow/component/add/graph-config';

export function FormatTime(time, formatStr = 'YYYY-MM-DD HH:mm:ss') {
  return moment(time).format(formatStr);
}

export function TimeDuration(start, end) {
  let duration = moment.duration(moment(end).diff(moment(start)));
  let seconds = duration.asSeconds();
  if (seconds >= 3600 * 24) {
    return '> 1d';
  }

  let hour = Math.floor(seconds / 3600);
  let minute = Math.floor((seconds % 3600) / 60);
  let second = seconds % 60;

  if (hour > 0) {
    return `${hour}h ${minute}m ${second}s`;
  } else if (minute > 0) {
    return `${minute}m ${second}s`;
  } else {
    return `${second}s`;
  }
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

export const formatTouchedField = value => {
  const touchObj = {};
  const flatObject = (obj, parent) => {
    _.forEach(obj, (v, k) => {
      if (_.isString(v)) {
        touchObj[`${parent}${k}`] = true;
      } else {
        flatObject(v, `${parent}${k}.`);
      }
    });
  };
  flatObject(value, '');
  return touchObj;
};
