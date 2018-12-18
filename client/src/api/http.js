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
const _message = mes => {
  if (mes === msg) {
    return;
  }
  message.error(mes, 1, () => {
    msg = '';
  });
  msg = mes;
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

//loading
let loading;
let loadingNum = 0;
const addLoading = function(method) {
  if (loadingNum <= 0) {
    loadingNum = 0;
    loading = message.loading(method === 'GET' ? '加载中' : '提交中', 3);
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
  baseURL: process.env.REACT_APP_BACKEND_SERVICE,
  timeout: 5000,
  headers: {
    accesstoken: getToken(),
    'Content-Type': 'application/json',
    'Access-Control-Allow-Origin': '*',
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
    addLoading(config.method.toUpperCase());
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
