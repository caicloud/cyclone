import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { Form, Radio, Select, Spin } from 'antd';
import { defaultFormItemLayout } from '@/lib/const';
import StageField from './StageField';

const Option = Select.Option;
const FormItem = Form.Item;

@inject('stageTemplate')
@observer
class AddStage extends React.Component {
  componentDidMount() {
    this.props.stageTemplate.getTemplateList();
  }
  state = {
    creationMethod: 'custom',
    templateData: null,
  };

  handleChange = e => {
    const value = e.target.value;
    this.setState({ creationMethod: value });
  };

  // Select template then render field
  selectTemplate = value => {
    const {
      stageTemplate: { templateList },
      setFieldValue,
      values,
    } = this.props;
    const item = _.find(
      _.get(templateList, 'items', []),
      o => _.get(o, 'metadata.name') === value
    );
    const inputs = _.merge({ name: '' }, _.get(item, 'spec.pod.inputs', {}));
    const currentStage = _.get(values, 'currentStage');
    setFieldValue(currentStage, inputs);
  };

  renderTemplateSelect = templates => {
    if (!templates) {
      return <Spin />;
    }

    const defaultValue = _.get(templates, '[0].metadata.name');
    return (
      <Select
        showSearch
        onSelect={this.selectTemplate}
        defaultValue={defaultValue}
      >
        {templates.map(o => {
          const name = _.get(o, 'metadata.name');
          return (
            <Option value={name} key={name}>
              {_.get(o, ['metadata', 'labels', 'cyclone.dev/builtin']) ===
              'true'
                ? intl.get(`template.kinds.${name.replace('-template', '')}`)
                : _.get(o, ['metadata', 'annotations', 'cyclone.dev/alias'])}
            </Option>
          );
        })}
      </Select>
    );
  };

  render() {
    const { creationMethod } = this.state;
    const {
      stageTemplate: { templateList },
    } = this.props;
    const templates = _.get(templateList, 'items');
    return (
      <Form>
        <FormItem
          label={intl.get('workflow.stageCreation')}
          {...defaultFormItemLayout}
        >
          <Radio.Group
            onChange={this.handleChange}
            defaultValue={creationMethod}
            value={creationMethod}
          >
            <Radio.Button value="template">
              {intl.get('sideNav.stageTemplate')}
            </Radio.Button>
            <Radio.Button value="custom">
              {intl.get('allocation.custom')}
            </Radio.Button>
          </Radio.Group>
        </FormItem>
        {creationMethod === 'template' && (
          <FormItem
            label={intl.get('workflow.selectTemplate')}
            {...defaultFormItemLayout}
          >
            {this.renderTemplateSelect(templates)}
          </FormItem>
        )}
        <StageField values={this.props.values} />
      </Form>
    );
  }
}

AddStage.propTypes = {
  stageTemplate: PropTypes.object,
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
};

export default AddStage;
