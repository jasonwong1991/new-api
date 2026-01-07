import { useState, useEffect } from 'react';
import { API, showError, showSuccess, copy } from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { Modal } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { useTableCompactMode } from '../common/useTableCompactMode';

export const INVITATION_STATUS = {
  ENABLED: 1,
  DISABLED: 2,
};

export const useInvitationsData = () => {
  const { t } = useTranslation();

  const [invitations, setInvitations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searching, setSearching] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [totalCount, setTotalCount] = useState(0);
  const [selectedKeys, setSelectedKeys] = useState([]);

  const [editingInvitation, setEditingInvitation] = useState({ id: undefined });
  const [showEdit, setShowEdit] = useState(false);
  const [showCreate, setShowCreate] = useState(false);

  const [formApi, setFormApi] = useState(null);
  const [compactMode, setCompactMode] = useTableCompactMode('invitations');

  const formInitValues = { searchKeyword: '' };

  const getFormValues = () => {
    const formValues = formApi ? formApi.getValues() : {};
    return { searchKeyword: formValues.searchKeyword || '' };
  };

  const loadInvitations = async (page = 1, size = pageSize) => {
    setLoading(true);
    try {
      const res = await API.get(`/api/invitation/?p=${page}&page_size=${size}`);
      const { success, message, data } = res.data;
      if (success) {
        setActivePage(data.page <= 0 ? 1 : data.page);
        setTotalCount(data.total);
        setInvitations(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const searchInvitations = async () => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadInvitations(1, pageSize);
      return;
    }

    setSearching(true);
    try {
      const res = await API.get(
        `/api/invitation/search?keyword=${searchKeyword}&p=1&page_size=${pageSize}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        setActivePage(data.page || 1);
        setTotalCount(data.total);
        setInvitations(data.items || []);
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setSearching(false);
  };

  const manageInvitation = async (id, action, record) => {
    setLoading(true);
    let res;
    try {
      switch (action) {
        case 'delete':
          res = await API.delete(`/api/invitation/${id}`);
          break;
        case 'enable':
          res = await API.put('/api/invitation/', { id, status: INVITATION_STATUS.ENABLED });
          break;
        case 'disable':
          res = await API.put('/api/invitation/', { id, status: INVITATION_STATUS.DISABLED });
          break;
        default:
          throw new Error('Unknown operation type');
      }

      const { success, message } = res.data;
      if (success) {
        showSuccess(t('操作成功'));
        await refresh();
      } else {
        showError(message);
      }
    } catch (error) {
      showError(error.message);
    }
    setLoading(false);
  };

  const refresh = async (page = activePage) => {
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      await loadInvitations(page, pageSize);
    } else {
      await searchInvitations();
    }
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadInvitations(page, pageSize);
    } else {
      searchInvitations();
    }
  };

  const handlePageSizeChange = (size) => {
    setPageSize(size);
    setActivePage(1);
    const { searchKeyword } = getFormValues();
    if (searchKeyword === '') {
      loadInvitations(1, size);
    } else {
      searchInvitations();
    }
  };

  const rowSelection = {
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const isExpired = (record) => {
    return (
      record.expired_time !== -1 &&
      record.expired_time < Math.floor(Date.now() / 1000)
    );
  };

  const isUsedUp = (record) => {
    return record.max_uses !== -1 && record.used_count >= record.max_uses;
  };

  const handleRow = (record) => {
    if (record.status !== INVITATION_STATUS.ENABLED || isExpired(record) || isUsedUp(record)) {
      return {
        style: { background: 'var(--semi-color-disabled-border)' },
      };
    }
    return {};
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板'));
    } else {
      Modal.error({
        title: t('无法复制到剪贴板，请手动复制'),
        content: text,
        size: 'large',
      });
    }
  };

  const batchCopyInvitations = async () => {
    if (selectedKeys.length === 0) {
      showError(t('请至少选择一个邀请码'));
      return;
    }
    let keys = selectedKeys.map((item) => item.code).join('\n');
    await copyText(keys);
  };

  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingInvitation({ id: undefined });
    }, 500);
  };

  const closeCreate = () => {
    setShowCreate(false);
  };

  useEffect(() => {
    loadInvitations(1, pageSize);
  }, [pageSize]);

  return {
    invitations,
    loading,
    searching,
    activePage,
    pageSize,
    totalCount,
    selectedKeys,

    editingInvitation,
    showEdit,
    showCreate,

    formApi,
    formInitValues,
    compactMode,
    setCompactMode,

    loadInvitations,
    searchInvitations,
    manageInvitation,
    refresh,
    copyText,

    setActivePage,
    setPageSize,
    setSelectedKeys,
    setEditingInvitation,
    setShowEdit,
    setShowCreate,
    setFormApi,
    setLoading,

    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    closeEdit,
    closeCreate,
    getFormValues,
    batchCopyInvitations,
    isExpired,
    isUsedUp,

    t,
  };
};
