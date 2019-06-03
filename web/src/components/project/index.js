import Project from './component/List';
import ProjectDetail from './component/detail';
import CreateProject from './component/addProject';
import CreateWorkflow from '@/components/workflow/component/add/Form';
import WorkflowDetail from '@/components/workflow/component/detail';
import { Route, Switch } from 'react-router-dom';
import PropTypes from 'prop-types';

const ProjectRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/projects" exact component={Project} />
      <Route
        path={`${match.path}/:projectName/workflows/add`}
        component={CreateWorkflow}
      />
      <Route
        path={`${match.path}/:projectName/workflows/:workflowName/update`}
        component={CreateWorkflow}
      />
      <Route
        path={`${match.path}/:projectName/workflows/:workflowName`}
        component={WorkflowDetail}
      />
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
