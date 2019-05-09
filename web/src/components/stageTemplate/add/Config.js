import PropTypes from 'prop-types';
import SectionCard from '@/components/public/sectionCard';
import MakeField from '@/components/public/makeField';
import { defaultFormItemLayout } from '@/lib/const';
import { Field, FieldArray } from 'formik';
import { Form, Input, Row, Col, Button } from 'antd';

const InputField = MakeField(Input);
const FormItem = Form.Item;
const Fragment = React.Fragment;

const ConfigSection = props => {
  const { values } = props;
  return (
    <SectionCard title={intl.get('config')}>
      <FieldArray
        name="spec.pod.inputs.arguments"
        render={() => (
          <Fragment>
            {_.get(values, 'spec.pod.inputs.arguments', []).map(
              (field, index) => (
                <Field
                  key={field.name}
                  label={intl.get(`template.form.config.${field.name}`)}
                  name={`spec.pod.inputs.arguments.${index}.value`}
                  component={InputField}
                  required
                />
              )
            )}
          </Fragment>
        )}
      />
      <FieldArray
        name="spec.pod.spec.containers"
        render={() => (
          <div>
            {_.get(values, 'spec.pod.spec.containers', []).map((a, index) => (
              <Fragment key={index}>
                <FormItem label={intl.get('env')} {...defaultFormItemLayout}>
                  <FieldArray
                    name={`spec.pod.spec.containers.${index}.env`}
                    render={arrayHelpers => (
                      <div>
                        {_.get(
                          values,
                          `spec.pod.spec.containers.${index}.env`,
                          []
                        ).length > 0 && (
                          <Row gutter={16}>
                            <Col span={11}>{intl.get('key')}</Col>
                            <Col span={11}>{intl.get('value')}</Col>
                          </Row>
                        )}
                        {_.get(
                          values,
                          `spec.pod.spec.containers.${index}.env`,
                          []
                        ).map((a, i) => (
                          <Row key={i} gutter={16}>
                            <Col span={11}>
                              <Field
                                key={a.name}
                                name={`spec.pod.spec.containers.${index}.env.${i}.name`}
                                component={InputField}
                                hasFeedback
                              />
                            </Col>
                            <Col span={11}>
                              <Field
                                key={a.value}
                                name={`spec.pod.spec.containers.${index}.env.${i}.value`}
                                component={InputField}
                                hasFeedback
                              />
                            </Col>
                            <Col span={2}>
                              <Button
                                type="circle"
                                icon="delete"
                                onClick={() => arrayHelpers.remove(i)}
                              />
                            </Col>
                          </Row>
                        ))}
                        <Button
                          ico="plus"
                          onClick={() => {
                            arrayHelpers.push({ name: '', value: '' });
                          }}
                        >
                          {intl.get('workflow.addEnv')}
                        </Button>
                      </div>
                    )}
                  />
                </FormItem>
              </Fragment>
            ))}
          </div>
        )}
      />
    </SectionCard>
  );
};

ConfigSection.propTypes = {
  values: PropTypes.object,
};

export default ConfigSection;
