import { Icon } from '@ui-kitten/components'
import React from 'react'
import { StyleSheet, View } from 'react-native'

import { UnifiedText } from '@berty/components/shared-components/UnifiedText'
import { useStyles } from '@berty/contexts/styles'

import { IMemberUserTypes } from './interfaces'

interface MemberNameProps extends IMemberUserTypes {
	displayName: string | null | undefined
}

const MemberUserTypes: React.FC<IMemberUserTypes> = ({ memberUserType = 'user' }) => {
	const { margin, text } = useStyles()

	let value, icon
	switch (memberUserType) {
		case 'replication':
			value = 'Replication node'
			icon = 'server'
			break
		case 'user':
			value = 'User device'
			icon = 'message-circle'
			break
	}

	return (
		<View style={[styles.container]}>
			<Icon name={icon} pack='feather' width={11} fill='#A8A8AA' />
			<UnifiedText style={[margin.left.scale(4), text.size.small, styles.text]}>
				{value}
			</UnifiedText>
		</View>
	)
}

export const MemberName: React.FC<MemberNameProps> = ({ displayName, memberUserType = 'user' }) => {
	return (
		<View>
			<UnifiedText>{displayName ?? ''}</UnifiedText>
			<MemberUserTypes memberUserType={memberUserType} />
		</View>
	)
}

const styles = StyleSheet.create({
	container: {
		flexDirection: 'row',
		alignItems: 'center',
	},
	text: { color: '#A8A8AA', fontFamily: 'Regular Open Sans' },
})
