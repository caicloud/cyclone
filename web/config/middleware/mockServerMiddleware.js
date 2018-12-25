const pathToRegexp = require('path-to-regexp');
const fs = require('fs');

// get actions to handle mock requests
function getMockActions(mockDir) {
  const files = fs.readdirSync(mockDir);
  const actions = {};
  // get all js files from mock dir as route controller
  const mods = files.filter(file => /(.*?)\.js/.test(file));
  mods.forEach(o => {
    const action = require(`${mockDir}/${o}`);
    for (const actionType in action) {
      actions[actionType] = action[actionType];
    }
  });
  return actions;
}

// Decode param value.
function decode_param(val) {
  if (typeof val !== 'string' || val.length === 0) {
    return val;
  }
  try {
    return decodeURIComponent(val);
  } catch (err) {
    return undefined;
  }
}

// get route params
function getParams(match, keys) {
  const params = {};
  for (let i = 1; i < match.length; i++) {
    const key = keys[i - 1];
    const prop = key.name;
    const val = decode_param(match[i]);

    if (
      val !== undefined ||
      !Object.prototype.hasOwnProperty.call(params, prop)
    ) {
      params[prop] = val;
    }
  }
  return params;
}

/**
 * match handler by actionName
 * @param {Object} actions
 * @param {String} actionName
 */
function matchAction(actions, actionName) {
  const result = { params: {}, handler: null };
  Object.keys(actions).forEach(handler => {
    const keys = [];
    const regexp = pathToRegexp(handler, keys);
    const match = regexp.exec(actionName);
    if (match) {
      result.params = getParams(match, keys);
      result.handler = handler;
    }
  });
  return result;
}

module.exports = function mockServerMiddleware(mockDir) {
  const mockActions = getMockActions(mockDir);

  return (req, res, next) => {
    // compose method & path as action name
    const actionName = `${req.method} ${req.path}`;
    const { params, handler } = matchAction(mockActions, actionName);
    const action = handler && mockActions[handler];

    if (typeof action === 'function') {
      req.params = Object.assign({}, req.param, params);
      return action(req, res);
    } else if (action) {
      // return data directly
      return res.json(action);
    }
    next();
  };
};
