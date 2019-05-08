import { Field } from 'formik';
import Selection from './Selection';
import GitHub from './Forms/GitHub';
import GitLab from './Forms/GitLab';
import PropTypes from 'prop-types';
import SVN from './Forms/SVN';
import Bitbucket from './Forms/Bitbucket';

const renderScmForm = (type, props) => {
  const ScmMap = {
    GitHub: <GitHub {...props} />,
    GitLab: <GitLab {...props} />,
    SVN: <SVN {...props} />,
    Bitbucket: <Bitbucket {...props} />,
  };
  return ScmMap[type];
};
const ScmGroup = props => {
  const {
    values: {
      spec: {
        scm: { type },
      },
    },
  } = props;
  return (
    <div className="u-scm">
      <h3>{intl.get('integration.form.scm.name')}</h3>
      <Field
        label={intl.get('type')}
        name="spec.scm.type"
        {...props}
        component={Selection}
      />
      <div style={{ margin: '24px 0' }}>
        {type && renderScmForm(type, props)}
      </div>
    </div>
  );
};

ScmGroup.propTypes = {
  values: PropTypes.object,
};

export default ScmGroup;
