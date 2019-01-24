import React from 'react';
import { BrowserRouter, Route } from 'react-router-dom';
import Integration from './component/List';
import AddSource from './component/addSource';

export default class Index extends React.Component {
  render() {
    return (
      <BrowserRouter>
        <div>
          <Route path="/integration" exact component={Integration} />
          <Route path="/integration/addsource" exact component={AddSource} />
        </div>
      </BrowserRouter>
    );
  }
}
