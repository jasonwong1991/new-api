import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import {
  API,
  downloadTextAsFile,
  showError,
  showSuccess,
} from '../../../../helpers';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import {
  Button,
  Modal,
  SideSheet,
  Space,
  Spin,
  Typography,
  Card,
  Tag,
  Form,
  Avatar,
  Row,
  Col,
} from '@douyinfe/semi-ui';
import {
  IconSave,
  IconClose,
  IconTicket,
} from '@douyinfe/semi-icons';

const { Text, Title } = Typography;

const EditInvitationModal = (props) => {
  const { t } = useTranslation();
  const isEdit = props.editingInvitation?.id !== undefined;
  const [loading, setLoading] = React.useState(isEdit);
  const isMobile = useIsMobile();
  const formApiRef = useRef(null);

  const getInitValues = () => ({
    name: '',
    count: 1,
    max_uses: 1,
    expired_time: null,
  });

  const handleCancel = () => {
    props.handleClose();
  };

  const loadInvitation = async () => {
    setLoading(true);
    let res = await API.get(`/api/invitation/${props.editingInvitation.id}`);
    const { success, message, data } = res.data;
    if (success) {
      const formData = { ...data };
      if (formData.expired_time === -1) {
        formData.expired_time = null;
      } else {
        formData.expired_time = new Date(formData.expired_time * 1000);
      }
      if (formData.max_uses === -1) {
        formData.max_uses = 0;
      }
      formApiRef.current?.setValues({ ...getInitValues(), ...formData });
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    if (formApiRef.current) {
      if (isEdit) {
        loadInvitation();
      } else {
        formApiRef.current.setValues(getInitValues());
      }
    }
  }, [props.editingInvitation?.id]);

  const submit = async (values) => {
    setLoading(true);
    let localInputs = { ...values };
    localInputs.count = parseInt(localInputs.count) || 1;
    localInputs.max_uses = parseInt(localInputs.max_uses) || 0;

    if (!localInputs.expired_time) {
      localInputs.expired_time = -1;
    } else {
      localInputs.expired_time = Math.floor(
        localInputs.expired_time.getTime() / 1000,
      );
    }

    if (localInputs.max_uses === 0) {
      localInputs.max_uses = -1;
    }

    let res;
    if (isEdit) {
      res = await API.put(`/api/invitation/`, {
        ...localInputs,
        id: parseInt(props.editingInvitation.id),
      });
    } else {
      res = await API.post(`/api/invitation/`, localInputs);
    }

    const { success, message, data } = res.data;
    if (success) {
      if (isEdit) {
        showSuccess(t('邀请码更新成功'));
        props.refresh();
        props.handleClose();
      } else {
        showSuccess(t('邀请码创建成功'));
        props.refresh();
        formApiRef.current?.setValues(getInitValues());
        props.handleClose();

        if (data && data.length > 0) {
          const codes = data.map((item) => item.code).join('\n');
          Modal.confirm({
            title: t('邀请码创建成功'),
            content: (
              <div>
                <p>{t('是否下载邀请码？')}</p>
                <p>{t('邀请码将以文本文件的形式下载。')}</p>
              </div>
            ),
            onOk: () => {
              downloadTextAsFile(codes, `invitation_codes_${localInputs.name || 'batch'}.txt`);
            },
          });
        }
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <SideSheet
      placement={isEdit ? 'right' : 'left'}
      title={
        <Space>
          {isEdit ? (
            <Tag color='blue' shape='circle'>
              {t('更新')}
            </Tag>
          ) : (
            <Tag color='green' shape='circle'>
              {t('新建')}
            </Tag>
          )}
          <Title heading={4} className='m-0'>
            {isEdit ? t('更新邀请码信息') : t('创建新的邀请码')}
          </Title>
        </Space>
      }
      bodyStyle={{ padding: '0' }}
      visible={props.visible}
      width={isMobile ? '100%' : 500}
      footer={
        <div className='flex justify-end bg-white'>
          <Space>
            <Button
              theme='solid'
              onClick={() => formApiRef.current?.submitForm()}
              icon={<IconSave />}
              loading={loading}
            >
              {t('提交')}
            </Button>
            <Button
              theme='light'
              type='primary'
              onClick={handleCancel}
              icon={<IconClose />}
            >
              {t('取消')}
            </Button>
          </Space>
        </div>
      }
      closeIcon={null}
      onCancel={handleCancel}
    >
      <Spin spinning={loading}>
        <Form
          initValues={getInitValues()}
          getFormApi={(api) => (formApiRef.current = api)}
          onSubmit={submit}
        >
          <div className='p-2'>
            <Card className='!rounded-2xl shadow-sm border-0'>
              <div className='flex items-center mb-2'>
                <Avatar size='small' color='blue' className='mr-2 shadow-md'>
                  <IconTicket size={16} />
                </Avatar>
                <div>
                  <Text className='text-lg font-medium'>
                    {t('邀请码设置')}
                  </Text>
                  <div className='text-xs text-gray-600'>
                    {t('设置邀请码的基本信息')}
                  </div>
                </div>
              </div>

              <Row gutter={12}>
                <Col span={24}>
                  <Form.Input
                    field='name'
                    label={t('名称')}
                    placeholder={t('请输入名称（可选）')}
                    style={{ width: '100%' }}
                    showClear
                  />
                </Col>
                {!isEdit && (
                  <Col span={12}>
                    <Form.InputNumber
                      field='count'
                      label={t('生成数量')}
                      min={1}
                      max={100}
                      rules={[
                        { required: true, message: t('请输入生成数量') },
                      ]}
                      style={{ width: '100%' }}
                      extraText={t('最多一次生成 100 个')}
                    />
                  </Col>
                )}
                <Col span={isEdit ? 24 : 12}>
                  <Form.InputNumber
                    field='max_uses'
                    label={t('最大使用次数')}
                    min={0}
                    style={{ width: '100%' }}
                    extraText={t('0 表示无限次数')}
                  />
                </Col>
                <Col span={24}>
                  <Form.DatePicker
                    field='expired_time'
                    label={t('过期时间')}
                    type='dateTime'
                    placeholder={t('选择过期时间（可选，留空为永久）')}
                    style={{ width: '100%' }}
                    showClear
                  />
                </Col>
              </Row>
            </Card>
          </div>
        </Form>
      </Spin>
    </SideSheet>
  );
};

export default EditInvitationModal;
