import { Field } from 'formik';
import MakeField from '@/components/public/makeField';
import { Input } from 'antd';

const InputField = MakeField(Input);

const DockerRegistry = () => {
  return (
    <div style={{ marginBottom: 24 }}>
      <h3>{intl.get('integration.form.dockerRegistry.name')}</h3>
      <Field
        label={intl.get('integration.form.dockerRegistry.registryAddress')}
        required
        name="spec.dockerRegistry.server"
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.dockerRegistry.admin')}
        required
        name="spec.dockerRegistry.user"
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.dockerRegistry.adminpwd')}
        required
        name="spec.dockerRegistry.password"
        component={InputField}
      />
    </div>
  );
};

export default DockerRegistry;
