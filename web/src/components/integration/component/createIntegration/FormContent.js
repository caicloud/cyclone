import { Field } from 'formik';
import PropTypes from 'prop-types';
import { Form, Input, Button, Row, Col } from 'antd';
import MakeField from '@/components/public/makeField';
import ScmGroup from '../formGroup/ScmGroup';
import SonarQube from '../formGroup/SonarQube';
import DockerRegistry from '../formGroup/DockerRegistry';
import Cluster from '../formGroup/Cluster';
import SelectSourceType from './SelectSourceType';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);
const renderWrapForm = (type, props) => {
  const formMap = {
    SCM: <ScmGroup {...props} />,
    DockerRegistry: <DockerRegistry {...props} />,
    SonarQube: <SonarQube {...props} />,
    Cluster: <Cluster {...props} />,
  };
  return formMap[type];
};
const SelectField = MakeField(SelectSourceType);
const FormContent = props => {
  const {
    values: {
      spec: { type },
    },
    setFieldValue,
    handleCancle,
    submit,
    update,
  } = props;
  return (
    <Form>
      <Field
        label={intl.get('integration.name')}
        name="metadata.alias"
        component={InputField}
        disabled={update}
        required
      />
      <Field
        label={intl.get('integration.desc')}
        name="metadata.description"
        component={TextareaField}
      />
      <Field
        label={intl.get('type')}
        name="spec.type"
        required
        disabled={update}
        handleSelectChange={val => {
          setFieldValue('spec.type', val);
        }}
        component={SelectField}
      />
      {type && renderWrapForm(type, props)}
      <Row>
        <Col offset={4} span={20}>
          <Button onClick={submit} type="primary">
            {intl.get('integration.form.confirm')}
          </Button>
          <Button style={{ marginLeft: 10 }} onClick={handleCancle}>
            {intl.get('integration.form.cancel')}
          </Button>
        </Col>
      </Row>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  values: PropTypes.object,
  submit: PropTypes.func,
  setFieldValue: PropTypes.func,
  handleCancle: PropTypes.func,
  update: PropTypes.bool,
};

export default FormContent;
