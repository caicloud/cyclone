import React from 'react';
import ReactDOM from 'react-dom';
import './less/index.less';
import store from './store';
import CoreLayout from './layout';
import { Provider } from 'mobx-react';
import registerServiceWorker from './registerServiceWorker';
import { BrowserRouter, Switch, Route } from 'react-router-dom';

ReactDOM.render(
  <Provider {...store}>
    <BrowserRouter>
      <Switch>
        <Route path="/" component={CoreLayout} />
      </Switch>
    </BrowserRouter>
  </Provider>,
  document.getElementById('root')
);
registerServiceWorker();
