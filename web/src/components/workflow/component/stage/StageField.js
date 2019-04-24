import { Input, Form, Button, Spin, Row, Col } from 'antd';
import MakeField from '@/components/public/makeField';
import { Field, FieldArray } from 'formik';
import SectionCard from '@/components/public/sectionCard';
import { defaultFormItemLayout } from '@/lib/const';
import Resource from '../resource/Form';
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
          <FormItem
            label={intl.get('sideNav.resource')}
            {...defaultFormItemLayout}
          >
            <FieldArray
              name={`${currentStage}.inputs.resources`}
              render={arrayHelpers => (
                <div>
                  {_.get(values, `${currentStage}.inputs.resources`, [])
                    .length > 0 && (
                    <Row gutter={16}>
                      <Col span={11}>{intl.get('name')}</Col>
                      <Col span={11}>{intl.get('path')}</Col>
                    </Row>
                  )}
                  {/* TODO(qme): click resource list show modal and restore resource form */}
                  {_.get(values, `${currentStage}.inputs.resources`, []).map(
                    (r, i) => (
                      <Row gutter={16} key={i}>
                        <Col span={11}>{r.name}</Col>
                        <Col span={11}>{r.path}</Col>
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
                  <Resource
                    SetReasourceValue={value => {
                      arrayHelpers.push(value);
                    }}
                  />
                </div>
              )}
            />
          </FormItem>
        </SectionCard>
        <SectionCard title={intl.get('config')}>
          <Field
            label={intl.get('image')}
            name={`${currentStage}.spec.containers.image`}
            component={InputField}
            hasFeedback
            required
          />
          <Field
            label={'ENTRYPOINT'}
            name={`${currentStage}.spec.containers.command`}
            component={TextareaField}
            hasFeedback
            required
          />
          <Field
            label={'ENTRYPOINT'}
            name={`${currentStage}.spec.containers.args`}
            component={TextareaField}
            hasFeedback
            required
          />
          <FormItem label={intl.get('env')} {...defaultFormItemLayout}>
            <FieldArray
              name={`${currentStage}.spec.containers.env`}
              render={arrayHelpers => (
                <div>
                  {_.get(values, `${currentStage}.spec.containers.env`, [])
                    .length > 0 && (
                    <Row gutter={16}>
                      <Col span={11}>{intl.get('key')}</Col>
                      <Col span={11}>{intl.get('value')}</Col>
                    </Row>
                  )}
                  {_.get(values, `${currentStage}.spec.containers.env`, []).map(
                    (a, i) => (
                      <Row key={i} gutter={16}>
                        <Col span={11}>
                          <Field
                            key={a.name}
                            name={`${currentStage}.spec.containers.env.${i}.name`}
                            component={InputField}
                            hasFeedback
                          />
                        </Col>
                        <Col span={11}>
                          <Field
                            key={a.value}
                            name={`${currentStage}.spec.containers.env.${i}.value`}
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
                    onClick={() => arrayHelpers.push({ name: '', value: '' })}
                  >
                    {intl.get('workflow.addEnv')}
                  </Button>
                </div>
              )}
            />
          </FormItem>
        </SectionCard>
        <SectionCard title={intl.get('output')}>
          <FormItem
            label={intl.get('sideNav.resource')}
            {...defaultFormItemLayout}
          >
            <FieldArray
              name={`${currentStage}.outputs.resources`}
              render={arrayHelpers => (
                <div>
                  {_.get(values, `${currentStage}.outputs.resources`, [])
                    .length > 0 && (
                    <Row gutter={16}>
                      <Col span={4}>{intl.get('name')}</Col>
                      <Col span={16}>{intl.get('image')}</Col>
                    </Row>
                  )}
                  {_.get(values, `${currentStage}.outputs.resources`, []).map(
                    (r, i) => {
                      const image = _.find(
                        _.get(r, 'spec.parameters'),
                        o => _.get(o, 'name') === 'IMAGE'
                      );
                      return (
                        <Row gutter={16} key={i}>
                          <Col span={4}>{r.name}</Col>
                          <Col span={16}>{_.get(image, 'value')}</Col>
                          <Col span={2}>
                            <Button
                              type="circle"
                              icon="delete"
                              onClick={() => arrayHelpers.remove(i)}
                            />
                          </Col>
                        </Row>
                      );
                    }
                  )}
                  <Resource
                    SetReasourceValue={value => {
                      arrayHelpers.push(value);
                    }}
                    type="output"
                  />
                </div>
              )}
            />
          </FormItem>
          <FormItem label={'artifacts'} {...defaultFormItemLayout}>
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
          </FormItem>
        </SectionCard>
      </Fragment>
    );
  }
}

export default StageField;
