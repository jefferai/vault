package pki

import (
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathRevoke(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `revoke`,
		Fields: map[string]*framework.FieldSchema{
			"serial_number": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Certificate serial number, in colon- or hyphen-separated octal",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.WriteOperation: b.pathRevokeWrite,
		},

		HelpSynopsis:    pathRevokeHelpSyn,
		HelpDescription: pathRevokeHelpDesc,
	}
}

func pathRotateCRL(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: `crl/rotate`,

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation: b.pathRotateCRLRead,
		},

		HelpSynopsis:    pathRotateCRLHelpSyn,
		HelpDescription: pathRotateCRLHelpDesc,
	}
}

func (b *backend) pathRevokeWrite(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	serial := data.Get("serial_number").(string)
	if len(serial) == 0 {
		return logical.ErrorResponse("The serial number must be provided"), nil
	}

	revokeStorageLock.Lock()
	defer revokeStorageLock.Unlock()

	return revokeCert(req, serial)
}

func (b *backend) pathRotateCRLRead(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	revokeStorageLock.Lock()
	defer revokeStorageLock.Unlock()

	userErr, intErr := buildCRL(req)
	switch {
	case userErr != nil:
		return logical.ErrorResponse(fmt.Sprintf("Error during CRL building: %s", userErr)), nil
	case intErr != nil:
		return nil, fmt.Errorf("Error encountered during CRL building: %s", intErr)
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"success": true,
		},
	}, nil
}

const pathRevokeHelpSyn = `
Revoke a certificate by serial number.
`

const pathRevokeHelpDesc = `
This allows certificates to be revoked using its serial number. A root token is required.
`

const pathRotateCRLHelpSyn = `
Force a rebuild of the CRL.
`

const pathRotateCRLHelpDesc = `
Force a rebuild of the CRL. This can be used to remove expired certificates from it if no certificates have been revoked. A root token is required.
`