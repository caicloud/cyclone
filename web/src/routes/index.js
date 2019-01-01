import { Switch, Route, Redirect } from 'react-router-dom';
import Project from '../components/project';
import Workflow from '../components/workflow';
import StageTemplate from '../components/stageTemplate';
import Resource from '../components/resource';
import Overview from '../components/overview';

import React from 'react';

const Routes = () => (
  <Switch>
    <Route path="/" exact component={Overview} />
    <Route path="/project" component={Project} />
    <Route path="/resource" component={Resource} />
    <Route path="/stageTemplate" component={StageTemplate} />
    <Route path="/workflow" component={Workflow} />
    <Redirect to="/" />
  </Switch>
);

export default Routes;
