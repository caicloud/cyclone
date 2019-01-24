import React from 'react';
import ReactDOM from 'react-dom';
import './less/index.less';
import store from './store';
import CoreLayout from './layout';
import { Provider } from 'mobx-react';
import { LocaleProvider } from 'antd';
import zhCN from 'antd/lib/locale-provider/zh_CN';
import enUS from 'antd/lib/locale-provider/en_US';
import _zhCN from './locale/zh-CN.yaml';
import _enUS from './locale/en-US.yaml';
import registerServiceWorker from './registerServiceWorker';
import { BrowserRouter, Switch, Route } from 'react-router-dom';

import intl from 'react-intl-universal';

const lang = localStorage.getItem('lang') || 'zh-CN';

intl.init({
  currentLocale: lang,
  locales: {
    [lang]: lang === 'zh-CN' ? _zhCN : _enUS,
  },
});

ReactDOM.render(
  <Provider {...store}>
    <BrowserRouter>
      <LocaleProvider locale={lang === 'zh-CN' ? zhCN : enUS}>
        <Switch>
          <Route path="/" component={CoreLayout} />
        </Switch>
      </LocaleProvider>
    </BrowserRouter>
  </Provider>,
  document.getElementById('root')
);
registerServiceWorker();
