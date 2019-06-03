import { Field } from 'formik';
import PropTypes from 'prop-types';
import { Form, Input, Button, Row, Col } from 'antd';
import MakeField from '@/components/public/makeField';

const { TextArea } = Input;
const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

const FormContent = ({
  history,
  handleSubmit,
  setFieldValue,
  update,
  submitCount,
}) => {
  return (
    <Form layout={'horizontal'} onSubmit={handleSubmit}>
      <Field
        label={intl.get('name')}
        name="metadata.alias"
        component={InputField}
        disabled={update}
        hasFeedback
        required
      />
      <Field
        label={intl.get('description')}
        name="metadata.description"
        component={TextareaField}
      />
      <Row>
        <Col span={4} />
        <Col span={12}>
          <div className="form-bottom-btn">
            <Button type="primary" htmlType="submit">
              {intl.get('operation.confirm')}
            </Button>
            <Button
              onClick={() => {
                history.push(`/projects`);
              }}
            >
              {intl.get('operation.cancel')}
            </Button>
          </div>
        </Col>
      </Row>
    </Form>
  );
};

FormContent.propTypes = {
  history: PropTypes.object,
  handleSubmit: PropTypes.func,
  setFieldValue: PropTypes.func,
  update: PropTypes.bool,
  submitCount: PropTypes.number,
};

export default FormContent;
