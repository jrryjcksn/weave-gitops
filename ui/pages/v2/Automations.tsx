import * as React from "react";
import styled from "styled-components";
import AutomationsTable from "../../components/AutomationsTable";
import Page from "../../components/Page";
import { AppContext } from "../../contexts/AppContext";
import { useListAutomations } from "../../hooks/automations";

type Props = {
  className?: string;
};

function Automations({ className }: Props) {
  const { api } = React.useContext(AppContext);
  const {
    data: automations,
    error,
    isLoading,
    isFetching,
  } = useListAutomations();

  React.useEffect(() => {
    api.ListNamespaces({});
  }, []);

  return (
    <Page
      error={error}
      loading={isLoading}
      title="Applications"
      isFetching={isFetching}
      className={className}
    >
      <AutomationsTable automations={automations} />
    </Page>
  );
}

export default styled(Automations).attrs({
  className: Automations.name,
})``;
