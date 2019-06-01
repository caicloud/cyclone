import { Form, Input } from 'antd';
import { Field, FieldArray } from 'formik';
import MakeField from '@/components/public/makeField';
import { noLabelItemLayout, modalFormItemLayout } from '@/lib/const';
import PropTypes from 'prop-types';
import { inject, observer } from 'mobx-react';
import { required } from '@/components/public/validate';
import SelectPlus from '@/components/public/makeField/select';

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
    bindInfo: PropTypes.object,
  };

  constructor(props) {
    super(props);
    const { bindInfo = {}, values } = props;
    this.state = {
      addWay: 'new',
      bindInfo,
      noShowIndex: this.getNoShowIndex(
        bindInfo,
        _.get(values, 'spec.parameters')
      ),
    };
  }

  getNoShowIndex = (bindInfo, specParameters) => {
    let noShowIndex = {};
    const noShowKey = _.keys(_.get(bindInfo, 'paramBindings'));
    _.forEach(specParameters, (v, i) => {
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
      this.setState({
        bindInfo: _.get(item, 'spec.bind'),
        noShowIndex: this.getNoShowIndex(
          _.get(item, 'spec.bind'),
          _.get(item, 'spec.parameters')
        ),
      });
      setFieldValue('spec.parameters', _.get(item, 'spec.parameters'));
    }
  };

  handleIntegrationChange = (val, list) => {
    const { bindInfo, noShowIndex } = this.state;
    const { setFieldValue } = this.props;
    setFieldValue('integration', val);
    const item = _.find(list, o => _.get(o, 'metadata.name'));
    if (item) {
      _.forEach(_.get(bindInfo, 'paramBindings'), (v, k) => {
        setFieldValue(
          `spec.parameters.${noShowIndex[k]}.value`,
          `$.${_.get(item, 'metadata.namespace')}.${val}/data.integration/${v}`
        );
      });
    }
  };

  renderParameters = noShowKeys => {
    const { values } = this.props;
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
            tooltip={field.description}
            validate={field.required && required}
          />
        );
        arr.push(dom);
      }
    });
    return arr;
  };

  render() {
    const { addWay, bindInfo } = this.state;
    const {
      values,
      setFieldValue,
      type,
      update,
      resource: { resourceTypeList },
      integration: { groupIntegrationList },
    } = this.props;
    const integrationResourceType = _.get(bindInfo, 'integrationType');
    const noShowKeys = _.keys(_.get(bindInfo, 'paramBindings'));
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
        ) : (
          <Fragment>
            {update ? (
              <FormItem label={intl.get('name')} {...modalFormItemLayout}>
                {_.get(values, 'name')}
              </FormItem>
            ) : (
              <Field
                label={intl.get('name')}
                name="name"
                component={InputField}
                formItemLayout={modalFormItemLayout}
                hasFeedback
                required
                validate={required}
              />
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
            <FormItem {...noLabelItemLayout}>
              <FieldArray
                name="spec.parameters"
                render={() => this.renderParameters(noShowKeys)}
              />
            </FormItem>
          </Fragment>
        )}
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
      </Form>
    );
  }
}

export default BindResource;
