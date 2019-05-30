import { Switch, Route, Redirect } from 'react-router-dom';
import Project from '../components/project';
import Workflow from '../components/workflow';
import StageTemplate from '../components/stageTemplate';
import Overview from '../components/overview';
import Resource from '../components/resource';
import Integration from '../components/integration';
import Swagger from '../components/swagger';
import React from 'react';

const Routes = () => (
  <Switch>
    <Route path="/" exact component={Overview} />
    <Route path="/project" component={Project} />
    <Route path="/stageTemplate" component={StageTemplate} />
    <Route path="/workflow" component={Workflow} />
    <Route path="/resource" component={Resource} />
    <Route path="/integration" component={Integration} />
    <Route path="/swagger" component={Swagger} />
    <Redirect to="/" />
  </Switch>
);

export default Routes;
