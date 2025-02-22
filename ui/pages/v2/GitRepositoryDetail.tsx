import * as React from "react";
import styled from "styled-components";
import Link from "../../components/Link";
import Page from "../../components/Page";
import SourceDetail from "../../components/SourceDetail";
import Timestamp from "../../components/Timestamp";
import {
  GitRepository,
  SourceRefSourceKind,
} from "../../lib/api/core/types.pb";

type Props = {
  className?: string;
  name: string;
  namespace: string;
};

function GitRepositoryDetail({ className, name, namespace }: Props) {
  return (
    <Page error={null} className={className} title={name}>
      <SourceDetail
        name={name}
        namespace={namespace}
        type={SourceRefSourceKind.GitRepository}
        info={(s: GitRepository) => [
          [
            "URL",
            <Link newTab href={s.url}>
              {s.url}
            </Link>,
          ],
          ["Ref", s.reference.branch],
          ["Last Updated", <Timestamp time={s.lastUpdatedAt} />],
          ["Cluster", s.clusterName],
          ["Namespace", s.namespace],
        ]}
      />
    </Page>
  );
}

export default styled(GitRepositoryDetail).attrs({
  className: GitRepositoryDetail.name,
})``;
