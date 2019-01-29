import { Route, Switch } from 'react-router-dom';
import Integration from './component/List';
import AddSource from './component/addSource';
import PropTypes from 'prop-types';

const IntegrationRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/integration" exact component={Integration} />
      <Route path={`${match.path}/add`} exact component={AddSource} />
    </Switch>
  );
};

IntegrationRoutes.propTypes = {
  match: PropTypes.object,
};

export default IntegrationRoutes;
