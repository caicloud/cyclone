import { POSITIVE_INT_OR_UP_TO_TWO_DIGITS_FLOAT } from '@/components/public/consts/regexp';

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
  const cpuLimit = _.get(value, 'limits.cpu');
  const memoryLimit = _.get(value, 'limits.memory');
  const cpuRequests = _.get(value, 'requests.cpu');
  const memoryRequests = _.get(value, 'requests.memory');

  const limitCpuErr = positiveOrFloat(cpuLimit);
  const limitMemErr = positiveOrFloat(memoryLimit);
  let reqCpuErr = positiveOrFloat(cpuRequests);
  let reqMemErr = positiveOrFloat(memoryRequests);

  if (value && value.limits && value.requests) {
    // TODO(qme): process unit conversion
    // if (Numeral.compare(value.requests.cpu, value.limits.cpu, 'cpu') > 0) {
    //   reqCpuErr = `CPU ${i18n.get('application.validates.quotaLimit')}`;
    // }
    // if (
    //   Numeral.compare(value.requests.memory, value.limits.memory, 'memory') > 0
    // ) {
    //   reqMemErr = `${i18n.get('all.memory')}${i18n.get(
    //     'application.validates.quotaLimit'
    //   )}`;
    // }
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
