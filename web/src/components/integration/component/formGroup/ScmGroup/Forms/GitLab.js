import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';
import AuthSelector from './AuthSelector';

const InputField = MakeField(Input);

const GitLab = props => {
  return (
    <div>
      <Field
        label={intl.get('integration.form.scm.serverAddress')}
        name="spec.scm.server"
        required
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.scm.authType')}
        {...props}
        name="spec.scm.authType"
        component={AuthSelector}
      />
    </div>
  );
};

export default GitLab;
