import { Route, Switch } from 'react-router-dom';
import CreateWorkFlow from './component/CreateWorkFlow';
import Workflow from './component/List';
import PropTypes from 'prop-types';

const WorkFlowRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/workflow" exact component={Workflow} />
      <Route path={`${match.path}/add`} exact component={CreateWorkFlow} />
    </Switch>
  );
};

WorkFlowRoutes.propTypes = {
  match: PropTypes.object,
};

export default WorkFlowRoutes;
