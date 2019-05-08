import { Field } from 'formik';
import { Input } from 'antd';
import MakeField from '@/components/public/makeField';
import AuthSelector from './AuthSelector';
import SwitchField from './SwitchField';
import PropTypes from 'prop-types';
import WorkerCluster from './WorkerCluster';

const InputField = MakeField(Input);

const Cluster = props => {
  const { setFieldValue } = props;
  return (
    <div style={{ marginBottom: 24 }}>
      <Field
        label={intl.get('integration.form.cluster.serverAddress')}
        name="spec.cluster.credential.server"
        required
        component={InputField}
      />
      <Field
        label={intl.get('integration.form.cluster.isControlCluster')}
        {...props}
        disabled
        onChange={val => {
          setFieldValue('spec.cluster.isControlCluster', val);
        }}
        name="spec.cluster.isControlCluster"
        component={SwitchField}
      />
      <WorkerCluster {...props} />
      <Field
        label={intl.get('integration.form.cluster.authType')}
        {...props}
        name="spec.cluster.credential.authType"
        component={AuthSelector}
      />
    </div>
  );
};

Cluster.propTypes = {
  setFieldValue: PropTypes.func,
};

export default Cluster;
