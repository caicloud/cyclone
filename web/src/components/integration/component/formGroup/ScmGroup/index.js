import { Field } from 'formik';
import Selection from './Selection';

const ScmGroup = props => {
  return (
    <div className="u-scm">
      <h3>{intl.get('integration.addexternalsystem')}</h3>
      <Field
        label={intl.get('integration.type')}
        name="spec.scm.type"
        {...props}
        component={Selection}
      />
    </div>
  );
};

export default ScmGroup;
