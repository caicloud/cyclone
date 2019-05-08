import { Route, Switch } from 'react-router-dom';
import CreateForm from './component/add/Form';
import Workflow from './component/list';
import PropTypes from 'prop-types';

const WorkFlowRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/workflow" exact component={Workflow} />
      <Route path={`${match.path}/add`} exact component={CreateForm} />
    </Switch>
  );
};

WorkFlowRoutes.propTypes = {
  match: PropTypes.object,
};

export default WorkFlowRoutes;
