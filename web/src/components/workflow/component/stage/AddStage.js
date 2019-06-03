import { inject, observer } from 'mobx-react';
import PropTypes from 'prop-types';
import { Form, Radio, Select, Spin } from 'antd';
import { drawerFormItemLayout, customStageField } from '@/lib/const';
import StageField from './StageField';
import TemplateStage from './TemplateStage';

const Option = Select.Option;
const FormItem = Form.Item;
const Fragment = React.Fragment;

@inject('stageTemplate')
@observer
class AddStage extends React.Component {
  constructor(props) {
    super(props);
    const { templateName, values } = props;
    const currentStage = _.get(values, 'currentStage');
    const stages = _.get(values, 'stages', []);
    const modify = stages.includes(currentStage);
    this.state = {
      creationMethod: modify && !templateName ? 'custom' : 'template',
      argDes: null,
      modify,
    };
  }

  componentDidMount() {
    const { update, values } = this.props;
    this.props.stageTemplate.getTemplateList(data => {
      if (update) {
        const stageTemplateName = _.get(values, [
          _.get(values, 'currentStage'),
          'metadata',
          'annotations',
          'stageTemplate',
        ]);
        const item = _.find(
          _.get(data, 'items', []),
          o => _.get(o, 'metadata.name') === stageTemplateName
        );
        const _arg = _.get(item, 'spec.pod.inputs.arguments', []);
        let argDes = {};
        _.forEach(_arg, v => {
          argDes[v.name] = v.description;
        });
        this.setState({ argDes });
      }
    });
  }

  handleChange = e => {
    const { setFieldValue, values } = this.props;
    const value = e.target.value;
    const state = { creationMethod: value };
    if (value === 'custom') {
      const currentStage = _.get(values, 'currentStage');
      const info = {
        metadata: { name: this.stageDefaultName() },
        ...customStageField,
      };
      setFieldValue(currentStage, info);
    } else {
      state.argDes = null;
    }
    this.setState(state);
  };

  transformTemplateData = (data, templateType) => {
    const value = _.cloneDeep(data);
    const _arguments = _.get(value, 'spec.pod.inputs.arguments', []);
    let argDes = {};

    _.forEach(_arguments, (v, k) => {
      if (v.name === 'config' && templateType === 'cd') {
        v.value = {
          namespace: '',
          name: '',
          container: '',
          image: '',
        };
      }
      if (v.name === 'cmd' && _.isString(v.value)) {
        const cmd = v.value.replace(/;\s+/g, ';\n');
        v.value = cmd;
      }
      argDes[v.name] = v.description;
      _arguments[k] = _.pick(v, ['name', 'value']);
    });

    const resource = _.concat(
      _.get(value, 'spec.pod.inputs.resources', []),
      _.get(value, 'spec.pod.outputs.resources', [])
    );
    _.forEach(resource, v => {
      if (v.name) {
        v.name = '';
      }
    });
    return { value, argDes };
  };

  stageDefaultName = () => {
    const { values } = this.props;
    const workflowName = _.get(values, 'metadata.name');
    const currentStageId = _.get(values, 'currentStage').split('_')[1];
    return `${workflowName}-stg${currentStageId}`;
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
    const templateType = _.get(item, [
      'metadata',
      'labels',
      'stage.cyclone.dev/template-kind',
    ]);
    let inputs = _.merge(
      {
        metadata: {
          name: this.stageDefaultName(),
          annotations: { stageTemplate: value },
        },
      },
      _.pick(item, 'spec', {})
    );
    if (templateType) {
      const formatData = this.transformTemplateData(inputs, templateType);
      inputs = _.get(formatData, 'value');
      this.setState({ argDes: formatData.argDes });
    }
    const currentStage = _.get(values, 'currentStage');
    setFieldValue(currentStage, inputs);
  };

  renderTemplateSelect = templates => {
    const { argDes } = this.state;
    if (!templates) {
      return <Spin />;
    }

    const defaultValue = _.get(templates, '[0].metadata.name');
    if (!argDes) {
      this.selectTemplate(defaultValue);
    }
    return (
      <Select
        showSearch
        onSelect={this.selectTemplate}
        defaultValue={defaultValue}
      >
        {templates.map(o => {
          const name = _.get(o, 'metadata.name');
          const namespace = _.get(o, 'metadata.namespace', '');
          const key = namespace ? `${namespace}-${name}` : name;
          return (
            <Option value={name} key={key}>
              {_.get(o, ['metadata', 'labels', 'cyclone.dev/builtin']) ===
              'true'
                ? intl.get(`template.kinds.${name.replace('-template', '')}`) ||
                  name
                : name}
            </Option>
          );
        })}
      </Select>
    );
  };

  render() {
    const { creationMethod, argDes, modify } = this.state;
    const {
      stageTemplate: { templateList },
      values,
      project,
      update,
      setFieldValue,
    } = this.props;
    const templates = _.get(templateList, 'items');
    const currentStage = _.get(values, 'currentStage');

    if (!_.get(values, `${currentStage}`)) {
      return <Spin />;
    }
    const stageProps = {
      values,
      update,
      modify,
      project,
      setFieldValue,
    };
    return (
      <Form>
        {!modify && (
          <FormItem
            label={intl.get('workflow.stageCreation')}
            {...drawerFormItemLayout}
          >
            <Radio.Group
              onChange={this.handleChange}
              defaultValue={creationMethod}
              value={creationMethod}
            >
              <Radio.Button value="template">
                {intl.get('sideNav.stageTemplate')}
              </Radio.Button>
              <Radio.Button value="custom">{intl.get('custom')}</Radio.Button>
            </Radio.Group>
          </FormItem>
        )}
        {creationMethod === 'template' ? (
          <Fragment>
            {!modify && (
              <FormItem
                label={intl.get('workflow.selectTemplate')}
                {...drawerFormItemLayout}
              >
                {this.renderTemplateSelect(templates)}
              </FormItem>
            )}
            <TemplateStage
              stageId={_.get(values, 'currentStage')}
              argDes={argDes}
              {...stageProps}
            />
          </Fragment>
        ) : (
          <StageField {...stageProps} />
        )}
      </Form>
    );
  }
}

AddStage.propTypes = {
  stageTemplate: PropTypes.object,
  setFieldValue: PropTypes.func,
  values: PropTypes.object,
  update: PropTypes.bool,
  project: PropTypes.string,
  templateName: PropTypes.string,
};

export default AddStage;
