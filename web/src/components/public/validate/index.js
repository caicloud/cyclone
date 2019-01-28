import { POSITIVE_INT_OR_UP_TO_TWO_DIGITS_FLOAT } from '@/consts/regexp';
import { parseQuantity as parse } from 'quantity.js';

export const required = val => {
  let empty = true;
  if (typeof val === 'string') {
    empty = !val.trim();
  } else if (_.isNumber(val)) {
    empty = false;
  } else {
    empty = _.isEmpty(val);
  }
  return empty ? intl.get('validate.required') : undefined;
};

// Positive number (including decimal, two decimal places at most)
export const positiveOrFloat = val => {
  const error = required(val);
  if (error) {
    return error;
  } else if (!POSITIVE_INT_OR_UP_TO_TWO_DIGITS_FLOAT.test(val)) {
    return intl.get('validate.positiveOrFloat');
  }
};

// resource validate, user for resourceAllocation component
/**
 * resource construction
 * @param {object} resources
 * {
 *    limits: {
 *      cpu,
 *      memroy,
 *    },
 *    requests: {
 *      cpu,
 *      memory,
 *    }
 * }
 */
export const resourceValidate = resources => {
  const err = {};
  const value = resources;
  const cpuLimit = _.replace(_.get(value, 'limits.cpu'), 'Core', '');
  const memoryLimit = _.replace(_.get(value, 'limits.memory'), /(M|G|T)i/g, '');
  const cpuRequests = _.replace(_.get(value, 'requests.cpu'), 'Core', '');
  const memoryRequests = _.replace(
    _.get(value, 'requests.memory'),
    /(M|G|T)i/g,
    ''
  );
  const limitCpuErr = positiveOrFloat(cpuLimit);
  const limitMemErr = positiveOrFloat(memoryLimit);
  let reqCpuErr = positiveOrFloat(cpuRequests);
  let reqMemErr = positiveOrFloat(memoryRequests);

  const convertMemoryLimit =
    _.get(value, 'limits.memory') &&
    parse(_.get(value, 'limits.memory'))
      .convertTo('Mi')
      .toString();
  const convertMemoryRequest =
    _.get(value, 'requests.memory') &&
    parse(_.get(value, 'requests.memory'))
      .convertTo('Mi')
      .toString();

  // TODO(qme): process unit conversion
  if (cpuRequests && cpuLimit && cpuRequests * 1 > cpuLimit * 1) {
    reqCpuErr = `CPU ${intl.get('validate.quota.exceedLimit')}`;
  }

  if (
    convertMemoryLimit &&
    convertMemoryRequest &&
    parseFloat(parse(convertMemoryRequest).minus(parse(convertMemoryLimit))) > 0
  ) {
    reqMemErr = `${intl.get('memory')}${intl.get(
      'validate.quota.exceedLimit'
    )}`;
  }

  if (limitCpuErr || limitMemErr) {
    err.limits = {
      cpu: limitCpuErr,
      memory: limitMemErr,
    };
  }
  if (reqCpuErr || reqMemErr) {
    err.requests = {
      cpu: reqCpuErr,
      memory: reqMemErr,
    };
  }

  if (_.isEmpty(err)) {
    return undefined;
  }
  return err;
};
