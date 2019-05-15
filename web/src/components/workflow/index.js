import { Route, Switch } from 'react-router-dom';
import CreateForm from './component/add/Form';
import Workflow from './component/list';
import WorkflowDetail from './component/detail';
import PropTypes from 'prop-types';

const WorkFlowRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/workflow" exact component={Workflow} />
      <Route path={`${match.path}/add`} component={CreateForm} />
      <Route
        path={`${match.path}/:workflowName/update`}
        component={CreateForm}
      />
      <Route path={`${match.path}/:workflowName`} component={WorkflowDetail} />
    </Switch>
  );
};

WorkFlowRoutes.propTypes = {
  match: PropTypes.object,
};

export default WorkFlowRoutes;
