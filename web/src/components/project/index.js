import Project from './component/List';
import ProjectDetail from './component/ProjectDetail';
import { Route, Switch } from 'react-router-dom';
import PropTypes from 'prop-types';
import React from 'react';

const ProjectRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/project" exact component={Project} />
      <Route path={`${match.path}/:projectId`} component={ProjectDetail} />
    </Switch>
  );
};

ProjectRoutes.propTypes = {
  match: PropTypes.object,
};
export default ProjectRoutes;
