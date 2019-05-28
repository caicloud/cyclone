import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import ResourceType from './ResourceType';
import MakeField from '@/components/public/makeField';
import { Field, FieldArray } from 'formik';
import { Form, Input, Row, Col, Button } from 'antd';
import { defaultFormItemLayout } from '@/lib/const';

const InputField = MakeField(Input);
const FormItem = Form.Item;

const InputSection = props => {
  const { values, setFieldValue, errors } = props;
  return (
    <SectionCard title={intl.get('input')}>
      <FormItem
        label={intl.get('template.form.inputs.arguments')}
        {...defaultFormItemLayout}
      >
        <FieldArray
          name="spec.pod.inputs.arguments"
          render={arrayHelpers => (
            <div>
              {_.get(values, 'spec.pod.inputs.arguments', []).length > 0 && (
                <Row gutter={16}>
                  <Col span={5}>name</Col>
                  <Col span={6}>value</Col>
                  <Col span={11}>desc</Col>
                </Row>
              )}
              {_.get(values, 'spec.pod.inputs.arguments', []).map(
                (a, index) => (
                  <Row key={index} gutter={16}>
                    <Col span={5}>
                      <Field
                        key={index}
                        name={`spec.pod.inputs.arguments.${index}.name`}
                        component={InputField}
                        hasFeedback
                      />
                    </Col>
                    <Col span={6}>
                      <Field
                        key={index}
                        name={`spec.pod.inputs.arguments.${index}.value`}
                        component={InputField}
                        hasFeedback
                      />
                    </Col>
                    <Col span={11}>
                      <Field
                        key={index}
                        name={`spec.pod.inputs.arguments.${index}.description`}
                        component={InputField}
                        hasFeedback
                      />
                    </Col>
                    <Col span={2}>
                      <Button
                        type="circle"
                        icon="delete"
                        onClick={() => arrayHelpers.remove(index)}
                      />
                    </Col>
                  </Row>
                )
              )}
              <Button
                ico="plus"
                onClick={() => {
                  arrayHelpers.push({ name: '', value: '', description: '' });
                }}
              >
                {intl.get('template.form.inputs.addArgs')}
              </Button>
            </div>
          )}
        />
      </FormItem>
      <ResourceType
        path="spec.pod.inputs.resources"
        values={values}
        setFieldValue={setFieldValue}
        errors={errors}
      />
    </SectionCard>
  );
};

InputSection.propTypes = {
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  errors: PropTypes.object,
};

export default InputSection;
