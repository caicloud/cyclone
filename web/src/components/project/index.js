import Project from './component/List';
import ProjectDetail from './component/detail';
import CreateProject from './component/addProject';
import { Route, Switch } from 'react-router-dom';
import PropTypes from 'prop-types';

const ProjectRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/project" exact component={Project} />
      <Route path={`${match.path}/add`} component={CreateProject} />
      <Route
        path={`${match.path}/:projectName/update`}
        component={CreateProject}
      />
      <Route path={`${match.path}/:projectName`} component={ProjectDetail} />
    </Switch>
  );
};

ProjectRoutes.propTypes = {
  match: PropTypes.object,
};
export default ProjectRoutes;
