import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { Input } from 'antd';

const InputField = MakeField(Input);

const SonarQube = () => {
  return (
    <div style={{ marginBottom: 24 }}>
      <h3>{intl.get('integration.form.sonarQube.name')}</h3>
      <Field
        label={intl.get('integration.form.sonarQube.serverAddress')}
        required
        name="spec.sonarQube.server"
        component={InputField}
      />
      <Field
        label="Token"
        required
        name="spec.sonarQube.token"
        component={InputField}
      />
    </div>
  );
};

export default SonarQube;
