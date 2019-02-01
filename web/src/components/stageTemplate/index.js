import List from './List/index';
import Detail from './Detail/index';
import { Route, Switch } from 'react-router-dom';
import PropTypes from 'prop-types';

const TemplateRoute = ({ match }) => {
  return (
    <Switch>
      <Route path="/stageTemplate" exact component={List} />
      <Route path={`${match.path}/:templateName`} component={Detail} />
    </Switch>
  );
};

TemplateRoute.propTypes = {
  match: PropTypes.object,
};
export default TemplateRoute;
