import React from 'react'
import { StyleSheet, TouchableOpacity } from 'react-native'

import beapi from '@berty/api'

import { GenericAvatar } from '../../avatars'
import { UnifiedText } from '../../shared-components/UnifiedText'
import { DropdownPriv } from '../Dropdown.priv'

interface AccountsDropdownProps {
	items: beapi.account.IAccountMetadata[]
	onChangeItem: (item: beapi.account.IAccountMetadata) => void
	placeholder: string
	defaultValue?: string | null
	testID?: string
}

export const AccountsDropdown: React.FC<AccountsDropdownProps> = props => {
	return (
		<DropdownPriv placeholder={props.placeholder} testID={props.testID}>
			{props.items.map((item, index) => (
				<TouchableOpacity
					onPress={() => props.onChangeItem(item)}
					style={[
						styles.item,
						{
							backgroundColor: props.defaultValue === item.accountId ? '#CFD2FB' : 'transparent',
							borderBottomLeftRadius: index === props.items.length - 1 ? 14 : 0,
							borderBottomRightRadius: index === props.items.length - 1 ? 14 : 0,
						},
					]}
					key={item.accountId}
				>
					<GenericAvatar size={25} colorSeed={item.publicKey} nameSeed={item.name} />
					<UnifiedText style={styles.text}>{item.name}</UnifiedText>
				</TouchableOpacity>
			))}
		</DropdownPriv>
	)
}

const styles = StyleSheet.create({
	item: {
		flexDirection: 'row',
		alignItems: 'center',
		paddingLeft: 32,
		paddingVertical: 12,
		paddingRight: 12,
		flex: 1,
	},
	text: {
		marginLeft: 7,
	},
})
