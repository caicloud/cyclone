import { Switch, Route, Redirect } from 'react-router-dom';
import Workspace from '../components/workspace';
import Pipeline from '../components/pipeline';
import React from 'react';

const Routes = () => (
  <Switch>
    <Route path="/" exact component={Workspace} />
    <Route path="/workspace" component={Workspace} />
    <Route path="/pipeline" component={Pipeline} />
    <Redirect to="/" />
  </Switch>
);

export default Routes;
