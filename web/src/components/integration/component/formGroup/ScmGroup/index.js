import { Field } from 'formik';
import Selection from './Selection';

const ScmGroup = props => {
  return (
    <div className="u-scm">
      <h3>{intl.get('integration.form.scm.codeOrigin')}</h3>
      <Field
        label={intl.get('type')}
        name="spec.scm.type"
        {...props}
        component={Selection}
      />
    </div>
  );
};

export default ScmGroup;
