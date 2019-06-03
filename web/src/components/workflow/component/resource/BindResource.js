import { Form, Input } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout, modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';
import SectionCard from '@/components/public/sectionCard';

const FormItem = Form.Item;
const InputField = MakeField(Input);
const SelectField = MakeField(SelectPlus);

const Fragment = React.Fragment;

// use in add stage, select a exist resource or create a new resource
@inject('integration', 'resource')
@observer
class BindResource extends React.Component {
  static propTypes = {
    setFieldValue: PropTypes.func,
    values: PropTypes.object,
    type: PropTypes.oneOf(['inputs', 'outputs']),
    integration: PropTypes.object,
    update: PropTypes.bool,
    resource: PropTypes.object,
    resourceTypeInfo: PropTypes.object,
  };

  constructor(props) {
    super(props);
    const { resourceTypeInfo = {} } = props;
    this.state = {
      addWay: 'new',
      resourceTypeInfo,
      noShowIndex: this.getNoShowIndex(resourceTypeInfo),
    };
  }

  getNoShowIndex = data => {
    let noShowIndex = {};
    const noShowKey = _.keys(_.get(data, 'bindInfo.paramBindings'));
    _.forEach(_.get(data, 'arg'), (v, i) => {
      if (noShowKey.includes(v.name)) {
        noShowIndex[v.name] = i;
      }
    });
    return noShowIndex;
  };

  componentDidMount() {
    const { integration } = this.props;
    integration.getIntegrationList();
  }

  changeAddWay = value => {
    this.setState({ addWay: value });
  };

  handleTypeChange = val => {
    const {
      setFieldValue,
      resource: { resourceTypeList },
    } = this.props;
    setFieldValue('type', val);
    setFieldValue('integration', '');
    const item = _.find(
      _.get(resourceTypeList, 'items', []),
      o => _.get(o, 'spec.type') === val
    );
    if (item) {
      const arg = _.get(item, 'spec.parameters', []);
      let argDes = {};
      _.forEach(arg, v => {
        argDes[v.name] = v.description;
      });
      const typeInfo = {
        bindInfo: _.get(item, 'spec.bind'),
        arg,
        argDes,
      };
      this.setState({
        resourceTypeInfo: typeInfo,
        noShowIndex: this.getNoShowIndex(typeInfo),
      });
      setFieldValue('spec.parameters', _.get(item, 'spec.parameters'));
    }
  };

  handleIntegrationChange = (val, list) => {
    const { resourceTypeInfo, noShowIndex } = this.state;
    const { setFieldValue } = this.props;
    setFieldValue('integration', val);
    const item = _.find(list, o => _.get(o, 'metadata.name'));
    if (item) {
      _.forEach(_.get(resourceTypeInfo, 'bindInfo.paramBindings'), (v, k) => {
        setFieldValue(
          `spec.parameters.${noShowIndex[k]}.value`,
          `$.${_.get(item, 'metadata.namespace')}.${val}/data.integration/${v}`
        );
      });
    }
  };

  renderParameters = noShowKeys => {
    const { values } = this.props;
    const { resourceTypeInfo } = this.state;
    let arr = [];
    _.forEach(_.get(values, 'spec.parameters', []), (field, index) => {
      if (!noShowKeys.includes(field.name)) {
        const dom = (
          <Field
            key={field.name}
            label={
              field.name.includes('SCM_')
                ? field.name.replace('SCM_', '')
                : field.name
            }
            name={`spec.parameters.${index}.value`}
            component={InputField}
            formItemLayout={modalFormItemLayout}
            hasFeedback
            required={field.required}
            tooltip={_.get(resourceTypeInfo, ['argDes', field.name])}
            validate={field.required && required}
          />
        );
        arr.push(dom);
      }
    });
    return arr;
  };

  render() {
    const { addWay, resourceTypeInfo } = this.state;
    const {
      values,
      setFieldValue,
      type,
      update,
      resource: { resourceTypeList },
      integration: { groupIntegrationList },
    } = this.props;
    const integrationResourceType = _.get(
      resourceTypeInfo,
      'bindInfo.integrationType'
    );
    const noShowKeys = _.keys(
      _.get(resourceTypeInfo, 'bindInfo.paramBindings')
    );
    const list = _.get(groupIntegrationList, integrationResourceType, []);
    return (
      <Form layout={'horizontal'}>
        <Field
          label={intl.get('workflow.resourceType')}
          name="type"
          handleSelectChange={val => {
            this.handleTypeChange(val);
          }}
          payload={{
            items: _.get(resourceTypeList, 'items', []),
            nameKey: 'spec.type',
            valueKey: 'spec.type',
          }}
          component={SelectField}
          formItemLayout={modalFormItemLayout}
          required
          validate={required}
        />
        {addWay === 'exist' ? (
          <Fragment>
            <Field
              label={intl.get('type')}
              name="name"
              handleSelectChange={val => {
                setFieldValue('name', val);
              }}
              component={<div>TODO: resource select</div>}
              formItemLayout={modalFormItemLayout}
              required
              validate={required}
            />
            {type === 'inputs' && (
              <Field
                label={intl.get('workflow.usePath')}
                name="path"
                component={InputField}
                hasFeedback
                required
                validate={required}
                formItemLayout={modalFormItemLayout}
              />
            )}
          </Fragment>
        ) : (
          <Fragment>
            {update ? (
              <Fragment>
                <FormItem label={intl.get('name')} {...modalFormItemLayout}>
                  {_.get(values, 'name')}
                </FormItem>
                {type === 'inputs' && (
                  <Field
                    label={intl.get('workflow.usePath')}
                    name="path"
                    component={InputField}
                    hasFeedback
                    required
                    validate={required}
                    formItemLayout={modalFormItemLayout}
                  />
                )}
              </Fragment>
            ) : (
              <Fragment>
                <Field
                  label={intl.get('name')}
                  name="name"
                  component={InputField}
                  formItemLayout={modalFormItemLayout}
                  hasFeedback
                  required
                  validate={required}
                />
                {type === 'inputs' && (
                  <Field
                    label={intl.get('workflow.usePath')}
                    name="path"
                    component={InputField}
                    hasFeedback
                    required
                    validate={required}
                    formItemLayout={modalFormItemLayout}
                  />
                )}
              </Fragment>
            )}
            <Field
              label={intl.get('sideNav.integration')}
              name="integration"
              payload={{
                items: list,
                nameKey: 'metadata.name',
                valueKey: 'metadata.name',
              }}
              handleSelectChange={val =>
                this.handleIntegrationChange(val, list)
              }
              component={SelectField}
              required
              validate={required}
              formItemLayout={modalFormItemLayout}
            />
            <SectionCard title={intl.get('resource.parameters')}>
              <FormItem {...noLabelItemLayout}>
                <FieldArray
                  name="spec.parameters"
                  render={() => this.renderParameters(noShowKeys)}
                />
              </FormItem>
            </SectionCard>
          </Fragment>
        )}
      </Form>
    );
  }
}

export default BindResource;
