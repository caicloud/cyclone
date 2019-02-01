import { Route, Switch } from 'react-router-dom';
import Integration from './component/List';
import CreateIntegration from './component/createIntegration';
import IntegrationDetail from './component/detail';
import PropTypes from 'prop-types';

const IntegrationRoutes = ({ match }) => {
  return (
    <Switch>
      <Route path="/integration" exact component={Integration} />
      <Route path={`${match.path}/add`} exact component={CreateIntegration} />
      <Route
        path={`${match.path}/:integrationName/update`}
        component={CreateIntegration}
      />
      <Route
        path={`${match.path}/:integrationName`}
        component={IntegrationDetail}
      />
    </Switch>
  );
};

IntegrationRoutes.propTypes = {
  match: PropTypes.object,
};

export default IntegrationRoutes;
