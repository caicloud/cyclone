import axios from 'axios';
import { message } from 'antd';
// getToken
const getToken = () => {
  if (localStorage.getItem('user')) {
    return JSON.parse(localStorage.getItem('user')).accessToken;
  }
  return '';
};

// hint
let msg = '';
const _message = m => {
  if (m === msg) {
    return;
  }
  message.error(m, 1, () => {
    msg = '';
  });
  msg = m;
};

// intercept repeated requests( rule: same url and same method)
const CancelToken = axios.CancelToken;
let requesting = {};
const cancelRequesting = config => {
  let requestId =
    config.method +
    config.url.replace(config.baseURL, '').replace(/^\//, '') +
    JSON.stringify(config.params);
  requesting[requestId] = false;
};
const addRequesting = config => {
  let cancel;
  config.cancelToken = new CancelToken(function executor(c) {
    cancel = c;
  });
  let requestId =
    config.method +
    config.url.replace(config.baseURL, '').replace(/^\//, '') +
    JSON.stringify(config.params);

  if (requesting[requestId]) {
    cancel({
      message: `重复请求`,
      config: config,
    });
    requesting[requestId] = false;
  } else {
    requesting[requestId] = true;
  }

  return config;
};

// get resource info from url
const getLoadingInfo = (method, url) => {
  const _querys = url.split('/');
  const subResource = ['stages', 'resources', 'workflows'];
  const firstResource = ['project', 'integrations', 'templates'];
  const includeSubResource = _.intersection(_querys, subResource);
  const includeFirstResource = _.intersection(_querys, firstResource);

  const lastResource =
    includeSubResource.length > 0 ? includeSubResource : includeFirstResource;
  let loadingText = '';
  if (lastResource && lastResource.length > 0) {
    const index = _.findIndex(_querys, o => o === lastResource[0]);
    loadingText = ['GET', 'POST'].includes(method)
      ? `${intl.get(method)} ${lastResource[0]}`
      : `${intl.get(method)} ${lastResource[0]} ${_querys[index + 1]}`;
  }
  return loadingText;
};

//loading
let loading;
let loadingNum = 0;
const addLoading = function(method, url) {
  if (loadingNum <= 0) {
    loadingNum = 0;
    loading = message.loading(getLoadingInfo(method, url), 3);
  }
  loadingNum++;
};
const removeLoading = function() {
  setTimeout(function() {
    loadingNum--;
    if (loading && loadingNum <= 0) {
      loading();
    }
  }, 0);
};
// axios instance
const instance = axios.create({
  baseURL: process.env.REACT_APP_API_BASE_URL,
  timeout: 5000,
  headers: {
    accesstoken: getToken(),
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
    'X-Tenant': 'system',
  },
  withCredentials: true,
});

instance.updateToken = token => {
  instance.defaults.headers['accesstoken'] = token || getToken();
};

// request interceptor
instance.interceptors.request.use(
  config => {
    if (!config.headers['accesstoken']) {
      instance.updateToken();
    }
    config = addRequesting(config);
    addLoading(config.method.toUpperCase(), config.url);
    return config;
  },
  function(error) {
    return Promise.reject(error);
  }
);

// response interceptor
instance.interceptors.response.use(
  response => {
    cancelRequesting(response.config);
    removeLoading();
    return Promise.resolve(response.data);
  },
  function(error) {
    if (error.response) {
      cancelRequesting(error.response.config);
    } else {
      requesting = {};
    }
    removeLoading();
    if (error.code === 'ECONNABORTED') {
      _message('网络连接超时');
    } else if (error.message === 'Request failed with status code 403') {
      _message('请重新登录');
      setTimeout(() => {
        window.location.href = window.location.origin + '/login';
      }, 1000);
    } else if (
      error.response &&
      error.response.data &&
      error.response.data.message
    ) {
      setTimeout(function() {
        _message(error.response.data.message);
      }, 300);
    }

    return Promise.reject(error);
  }
);

export default instance;
