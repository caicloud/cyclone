import { Input, Form, Button, Spin, Row, Col } from 'antd';
import MakeField from '@/components/public/makeField';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import { defaultFormItemLayout } from '@/lib/const';
import ResourceArray from '../resource/ResourceArray';
import PropTypes from 'prop-types';

const Fragment = React.Fragment;
const FormItem = Form.Item;
const { TextArea } = Input;

const InputField = MakeField(Input);
const TextareaField = MakeField(TextArea);

class StageField extends React.Component {
  static propTypes = {
    values: PropTypes.object,
  };

  state = {
    visible: false,
  };

  render() {
    const { values } = this.props;
    const currentStage = _.get(values, 'currentStage');
    if (!currentStage) {
      return <Spin />;
    }
    return (
      <Fragment>
        <Field
          label={intl.get('name')}
          name={`${currentStage}.name`}
          component={InputField}
          hasFeedback
          required
        />
        <SectionCard title={intl.get('input')}>
          <ResourceArray
            resourcesField={`${currentStage}.inputs.resources`}
            resources={_.get(values, `${currentStage}.inputs.resources`, [])}
          />
        </SectionCard>
        <SectionCard title={intl.get('config')}>
          <FieldArray
            name={`${currentStage}.spec.containers`}
            render={() => (
              <div>
                {_.get(values, `${currentStage}.spec.containers`, []).map(
                  (a, index) => (
                    <Fragment key={index}>
                      <Field
                        label={intl.get('image')}
                        name={`${currentStage}.spec.containers.${index}.image`}
                        component={InputField}
                        hasFeedback
                        required
                      />
                      <Field
                        label={'ENTRYPOINT'}
                        name={`${currentStage}.spec.containers.${index}.command`}
                        component={TextareaField}
                        hasFeedback
                        required
                      />
                      <Field
                        label={'COMMAND'}
                        name={`${currentStage}.spec.containers.${index}.args`}
                        component={TextareaField}
                        hasFeedback
                        required
                      />
                      <FormItem
                        label={intl.get('env')}
                        {...defaultFormItemLayout}
                      >
                        <FieldArray
                          name={`${currentStage}.spec.containers.${index}.env`}
                          render={arrayHelpers => (
                            <div>
                              {_.get(
                                values,
                                `${currentStage}.spec.containers.${index}.env`,
                                []
                              ).length > 0 && (
                                <Row gutter={16}>
                                  <Col span={11}>{intl.get('key')}</Col>
                                  <Col span={11}>{intl.get('value')}</Col>
                                </Row>
                              )}
                              {_.get(
                                values,
                                `${currentStage}.spec.containers.${index}.env`,
                                []
                              ).map((a, i) => (
                                <Row key={i} gutter={16}>
                                  <Col span={11}>
                                    <Field
                                      key={a.name}
                                      name={`${currentStage}.spec.containers.${index}.env.${i}.name`}
                                      component={InputField}
                                      hasFeedback
                                    />
                                  </Col>
                                  <Col span={11}>
                                    <Field
                                      key={a.value}
                                      name={`${currentStage}.spec.containers.${index}.env.${i}.value`}
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
                                onClick={() =>
                                  arrayHelpers.push({ name: '', value: '' })
                                }
                              >
                                {intl.get('workflow.addEnv')}
                              </Button>
                            </div>
                          )}
                        />
                      </FormItem>
                    </Fragment>
                  )
                )}
              </div>
            )}
          />
        </SectionCard>
        <SectionCard title={intl.get('output')}>
          <ResourceArray
            type="outputs"
            resourcesField={`${currentStage}.outputs.resources`}
            resources={_.get(values, `${currentStage}.outputs.resources`, [])}
          />
          {/* // NOTE: temporarily not supported artifacts */}
          {/* <FormItem label={'artifacts'} {...defaultFormItemLayout}>
            <FieldArray
              name={`${currentStage}.outputs.artifacts`}
              render={arrayHelpers => (
                <div>
                  {_.get(values, `${currentStage}.outputs.artifacts`, [])
                    .length > 0 && (
                    <Row gutter={16}>
                      <Col span={11}>{intl.get('name')}</Col>
                      <Col span={11}>{intl.get('path')}</Col>
                    </Row>
                  )}
                  {_.get(values, `${currentStage}.outputs.artifacts`, []).map(
                    (r, i) => (
                      <Row key={i} gutter={16}>
                        <Col span={11}>
                          <Field
                            key={r.name}
                            name={`${currentStage}.outputs.artifacts.${i}.name`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        <Col span={11}>
                          <Field
                            key={r.name}
                            name={`${currentStage}.outputs.artifacts.${i}.path`}
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
                    )
                  )}
                  <Button
                    ico="plus"
                    onClick={() => arrayHelpers.push({ name: '', path: '' })}
                  >
                    {intl.get('workflow.addArtifact')}
                  </Button>
                </div>
              )}
            />
          </FormItem> */}
        </SectionCard>
      </Fragment>
    );
  }
}

export default StageField;
